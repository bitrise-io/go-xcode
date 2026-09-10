package errorfinder

import (
	"bytes"
	"strings"
	"sync"
)

// maxLineLength bounds how much of a line is held back while waiting for its newline. Error lines are short,
// so a longer line is not one: it is skipped instead of buffered.
const maxLineLength = 1024 * 1024

// Finder finds xcodebuild errors in a stream of log output.
//
// It is an io.Writer so that it can be teed off a log pipeline, and the writes do not have to be line aligned:
// a partial line is held back until its newline arrives. Call Errors once the stream ended.
type Finder struct {
	mu sync.Mutex

	pending  []byte // partial trailing line, held until its newline arrives
	skipping bool   // the pending line grew past maxLineLength, the rest of it is dropped

	errorLines        []string  // single line errors with "error: " prefix
	xcodebuildErrors  []string  // multiline errors starting with "xcodebuild: error: " prefix
	nserrors          []nsError // single line NSErrors with schema: Error Domain=<domain> Code=<code> "<reason>" UserInfo=<user_info>
	isXcodebuildError bool
	xcodebuildError   string
}

// NewFinder returns a Finder ready to receive output.
func NewFinder() *Finder {
	return &Finder{}
}

// Write consumes the next chunk of output. It never fails and always reports the whole chunk as written, so it
// can not interrupt an io.MultiWriter it is part of.
func (f *Finder) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	n := len(p)
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			f.appendPending(p)
			break
		}
		f.appendPending(p[:i])
		f.finishLine()
		p = p[i+1:]
	}

	return n, nil
}

// Errors returns the errors found so far. The current end of the output is treated as the end of a line (a last
// line without a newline is processed, an open multiline error is closed), so it is safe to call more than once.
func (f *Finder) Errors() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.pending) > 0 || f.skipping {
		f.finishLine()
	}
	f.closeXcodebuildError()

	// Regular error lines (line with 'error: ' prefix) seems to have
	// NSError line pairs (same description) in some cases.
	errorLines := intersection(f.errorLines, f.nserrors)

	return append(errorLines, f.xcodebuildErrors...)
}

// FindXcodebuildErrors returns the xcodebuild errors found in out. See Finder for the streaming form.
func FindXcodebuildErrors(out string) []string {
	f := NewFinder()
	_, _ = f.Write([]byte(out))

	return f.Errors()
}

func (f *Finder) appendPending(b []byte) {
	if f.skipping {
		return
	}
	if len(f.pending)+len(b) > maxLineLength {
		f.pending = f.pending[:0]
		f.skipping = true

		return
	}
	f.pending = append(f.pending, b...)
}

// finishLine feeds the completed pending line to the state machine. A skipped (overlong) line is not an error
// line, but it still ends a multiline xcodebuild error, so it is consumed as an empty line.
func (f *Finder) finishLine() {
	line := ""
	if !f.skipping {
		line = string(f.pending)
	}
	f.pending = f.pending[:0]
	f.skipping = false

	f.consumeLine(line)
}

// consumeLine is one step of the line-oriented state machine.
func (f *Finder) consumeLine(line string) {
	if f.isXcodebuildError {
		line = strings.TrimLeft(line, " ")
		if strings.HasPrefix(line, "Reason: ") || strings.HasPrefix(line, "Recovery suggestion: ") {
			f.xcodebuildError += "\n" + line

			return
		}
		f.closeXcodebuildError()
	}

	switch {
	case strings.HasPrefix(line, "xcodebuild: error: "):
		f.xcodebuildError = line
		f.isXcodebuildError = true
	case strings.HasPrefix(line, "error: ") || strings.Contains(line, " error: "):
		f.errorLines = append(f.errorLines, line)
	case strings.HasPrefix(line, "Error "):
		if e := newNSError(line); e != nil {
			f.nserrors = append(f.nserrors, *e)
		}
	}
}

func (f *Finder) closeXcodebuildError() {
	if !f.isXcodebuildError {
		return
	}
	f.xcodebuildErrors = append(f.xcodebuildErrors, f.xcodebuildError)
	f.xcodebuildError = ""
	f.isXcodebuildError = false
}

func intersection(errorLines []string, nserrors []nsError) []string {
	union := make([]string, len(errorLines))
	copy(union, errorLines)

	for _, nserror := range nserrors {
		found := false
		for i, errorLine := range errorLines {
			// Checking suffix, as regular error lines have additional prefixes, like "error: exportArchive: "
			if strings.HasSuffix(errorLine, nserror.Description) {
				union[i] = nserror.Error()
				found = true
				break
			}
		}
		if !found {
			union = append(union, nserror.Error())
		}
	}
	return union
}
