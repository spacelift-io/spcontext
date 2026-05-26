package spcontext

import (
	"runtime"
	"strings"

	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/pkg/errors"
)

// BugsnagLogger wraps the given Context inside a bugsnag friendly logger.
type BugsnagLogger struct {
	Ctx Context
}

// Printf logs the message with info level.
func (l *BugsnagLogger) Printf(format string, v ...interface{}) {
	l.Ctx.Infof(format, v...)
}

type stackTracer interface {
	error
	StackTrace() errors.StackTrace
}

// errorWithStackFrames satisfies bugsnag.ErrorWithStackFrames for a github.com/pkg/errors error.
type errorWithStackFrames struct {
	err stackTracer
}

// Cause returns initial error, provides compatibility for pkg/errors chains.
func (w *errorWithStackFrames) Cause() error { return w.err }

// Unwrap returns initial error, provides compatibility for Go 1.13 error chains.
func (w *errorWithStackFrames) Unwrap() error { return w.err }

func (e *errorWithStackFrames) Error() string {
	return e.err.Error()
}

// hasInitTimeStackOnly reports whether st's stack contains only frames from
// process startup (runtime.*, main.main, package init functions).
//
// This is the signature of a pkg/errors error created at package init via
// `var ErrFoo = errors.New("...")` or similar, whose captured stack points to nothing
// meaningful.
//
// The chain walk in (*Context).error skips such stacks so Bugsnag receives a useful
// frame instead.
func hasInitTimeStackOnly(st stackTracer) bool {
	trace := st.StackTrace()
	if len(trace) == 0 {
		return false
	}
	for _, frame := range trace {
		pc := uintptr(frame) - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			return false
		}
		if !isInitOrRuntimeFrame(fn.Name()) {
			return false
		}
	}
	return true
}

// isInitOrRuntimeFrame matches function names produced by the Go runtime's
// startup sequence and compiler-generated package init functions.
func isInitOrRuntimeFrame(name string) bool {
	switch {
	case strings.HasPrefix(name, "runtime."):
		return true
	case name == "main.main":
		return true
	case strings.HasSuffix(name, ".init"):
		return true
	case strings.Contains(name, ".init."):
		return true
	}
	return false
}

func (e *errorWithStackFrames) StackFrames() []bugsnagerrors.StackFrame {
	stackTrace := e.err.StackTrace()

	out := make([]bugsnagerrors.StackFrame, len(stackTrace))
	for i, frame := range stackTrace {
		pc := uintptr(frame) - 1
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)

		name := fn.Name()
		var pkg string
		var fnName string
		if pkgDivider := strings.LastIndex(name, "/"); pkgDivider != -1 {
			pkg = name[:pkgDivider]
			fnName = name[pkgDivider+1:]
			if fnDivider := strings.LastIndex(fnName, "."); fnDivider != -1 {
				fnName = fnName[fnDivider+1:]
			}
		} else if pkgDivider := strings.Index(name, "."); pkgDivider != -1 {
			pkg = name[:pkgDivider]
			fnName = name[pkgDivider+1:]
			if fnDivider := strings.LastIndex(fnName, "."); fnDivider != -1 {
				fnName = fnName[fnDivider+1:]
			}
		}

		out[i] = bugsnagerrors.StackFrame{
			File:           file,
			LineNumber:     line,
			Name:           fnName,
			Package:        pkg,
			ProgramCounter: pc,
		}
	}

	return out
}
