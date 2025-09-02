package loginterceptor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
)

// PrefixInterceptor intercept writes: if a line begins with prefix, it will be written to
// both writers. Partial writes without newline are buffered until a newline.
type PrefixInterceptor struct {
	re          *regexp.Regexp
	intercepted io.Writer
	original    io.Writer

	// internal pipe and goroutine to scan and route
	pr *io.PipeReader
	pw *io.PipeWriter

	// close once
	closeOnce sync.Once
	closeErr  error
}

// NewPrefixInterceptor returns an io.WriteCloser. Writes are based on line prefix.
func NewPrefixInterceptor(re *regexp.Regexp, intercepted, original io.Writer) *PrefixInterceptor {
	pr, pw := io.Pipe()
	interceptor := &PrefixInterceptor{
		re:          re,
		intercepted: intercepted,
		original:    original,
		pr:          pr,
		pw:          pw,
	}
	go interceptor.run()
	return interceptor
}

// Write implements io.Writer. It writes into an internal pipe which the interceptor goroutine consumes.
func (i *PrefixInterceptor) Write(p []byte) (int, error) {
	return i.pw.Write(p)
}

// Close stops the interceptor and closes the pipe.
func (i *PrefixInterceptor) Close() error {
	i.closeOnce.Do(func() {
		// close the writer side which causes reader side to EOF
		i.closeErr = i.pw.Close()
		// ensure reader is drained (run goroutine will finish)
		_ = i.pr.Close()
	})
	return i.closeErr
}

// run reads lines (and partial final chunk) and writes them.
func (i *PrefixInterceptor) run() {
	defer func() {
		// Ensure pipe reader is closed when goroutine exits
		_ = i.pr.Close()
	}()

	// Use a scanner but with a large buffer to handle long lines.
	scanner := bufio.NewScanner(i.pr)
	const maxTokenSize = 10 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text() // note: newline removed
		// re-append newline to preserve same output format
		outLine := line + "\n"
		if i.re.MatchString(line) {
			if _, err := io.WriteString(i.intercepted, outLine); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "intercepted writer error: %v\n", err)
			}
		}
		if _, err := io.WriteString(i.original, outLine); err != nil {
			// Log error but continue processing
			_, _ = fmt.Fprintf(os.Stderr, "original writer error: %v\n", err)
		}
	}
	// handle any scanner error
	if err := scanner.Err(); err != nil && err != io.EOF {
		_, _ = fmt.Fprintf(os.Stderr, "router scanner error: %v\n", err)
	}
}
