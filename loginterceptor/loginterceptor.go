package loginterceptor

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
)

// PrefixInterceptor intercept writes: if a line begins with prefix, it will be written to
// both writers. Partial writes without newline are buffered until a newline.
type PrefixInterceptor struct {
	prefixRegexp *regexp.Regexp
	intercepted  io.Writer
	target       io.Writer
	logger       log.Logger

	// internal pipe and goroutine to scan and route
	internalReader *io.PipeReader
	internalWriter *io.PipeWriter
	writeTimeout   time.Duration

	// close once
	closeOnce sync.Once
	closeErr  error
}

// NewPrefixInterceptor returns an io.WriteCloser. Writes are based on line prefix.
func NewPrefixInterceptor(prefixRegexp *regexp.Regexp, intercepted, target io.Writer, logger log.Logger) *PrefixInterceptor {
	return NewPrefixInterceptorWithTimeout(prefixRegexp, intercepted, target, logger, 1*time.Second)
}

// NewPrefixInterceptorWithTimeout returns an io.WriteCloser. Writes are based on line prefix.
func NewPrefixInterceptorWithTimeout(prefixRegexp *regexp.Regexp, intercepted, target io.Writer, logger log.Logger, writeTimeout time.Duration) *PrefixInterceptor {
	pipeReader, pipeWriter := io.Pipe()
	interceptor := &PrefixInterceptor{
		prefixRegexp:   prefixRegexp,
		intercepted:    intercepted,
		target:         target,
		logger:         logger,
		internalReader: pipeReader,
		internalWriter: pipeWriter,
		writeTimeout:   writeTimeout,
	}

	interceptCh := WriterWorker(interceptor.intercepted)
	targetCh := WriterWorker(interceptor.target)

	go interceptor.scan(interceptCh, targetCh)

	return interceptor
}

// Write implements io.Writer. It writes into an internal pipe which the interceptor goroutine consumes.
func (i *PrefixInterceptor) Write(p []byte) (int, error) {
	return i.internalWriter.Write(p)
}

// Close stops the interceptor and closes the pipe.
func (i *PrefixInterceptor) Close() error {
	i.closeOnce.Do(func() {
		i.closeErr = i.internalWriter.Close()
	})

	return i.closeErr
}

func (i *PrefixInterceptor) scan(reqChIntercepted, reqChTarget chan<- writeReq) {
	defer func() {
		if err := i.internalReader.Close(); err != nil {
			i.logger.Errorf("internal reader: %v", err)
		}
		// Close writers if able
		if interceptedCloser, ok := i.intercepted.(io.Closer); ok {
			if err := interceptedCloser.Close(); err != nil {
				i.logger.Errorf("closing intercepted writer: %v", err)
			}
		}
		if originalCloser, ok := i.target.(io.Closer); ok {
			if err := originalCloser.Close(); err != nil {
				i.logger.Errorf("closing original writer: %v", err)
			}
		}
		defer close(reqChIntercepted) // stop the worker when done
		defer close(reqChTarget)      // stop the worker when done
	}()

	// Use a scanner but with a large buffer to handle long lines.
	scanner := bufio.NewScanner(i.internalReader)
	const maxTokenSize = 10 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text() // note: newline removed
		// re-append newline to preserve same output format
		msg := line + "\n"

		if i.prefixRegexp.MatchString(msg) {
			ctx, cancel := context.WithTimeout(context.Background(), i.writeTimeout)
			if _, err := WriteWithContext(ctx, reqChIntercepted, []byte(msg)); err != nil {
				i.logger.Errorf("intercept writer error: %v", err)
			}
			cancel()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if _, err := WriteWithContext(ctx, reqChTarget, []byte(msg)); err != nil {
			i.logger.Errorf("writer error: %v", err)
		}
		cancel()
	}

	// handle any scanner error
	if err := scanner.Err(); err != nil {
		i.logger.Errorf("interceptor scanner error: %v\n", err)
	}
}

// writeReq represents a single write request.
type writeReq struct {
	ctx  context.Context
	data []byte
	resp chan writeResp
}

type writeResp struct {
	n   int
	err error
}

// WriterWorker serializes writes to w. It returns a channel to submit requests.
// Close the returned channel to stop the worker (it will finish current request and exit).
func WriterWorker(w io.Writer) chan<- writeReq {
	reqCh := make(chan writeReq)
	go func() {
		for req := range reqCh {
			// perform the blocking write in a helper goroutine so we can honor req.ctx
			done := make(chan writeResp, 1)

			go func(r writeReq) {
				n, err := w.Write(r.data)
				done <- writeResp{n: n, err: err}
			}(req)

			select {
			case <-req.ctx.Done():
				// Caller gave up. We still need to wait for the underlying write goroutine
				// to finish to avoid leaking it, but we return the ctx error to caller.
				// Option A: just wait and discard result
				go func() {
					<-done // let the write goroutine finish and drop the result
				}()
				req.resp <- writeResp{n: 0, err: req.ctx.Err()}
			case res := <-done:
				// Write finished before context done.
				req.resp <- res
			}
			close(req.resp)
		}
	}()
	return reqCh
}

// WriteWithContext is a helper to submit a write and wait respecting ctx
func WriteWithContext(ctx context.Context, reqCh chan<- writeReq, data []byte) (int, error) {
	resp := make(chan writeResp, 1)
	select {
	case reqCh <- writeReq{ctx: ctx, data: data, resp: resp}:
		// request submitted
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	select {
	case r := <-resp:
		return r.n, r.err
	case <-ctx.Done():
		// If context expires after the request was submitted but before we read the resp,
		// we still return ctx.Err(). The worker side will also return ctx.Err. to resp
		// in the previous select if it sees ctx.Done() before write completes.
		return 0, ctx.Err()
	}
}
