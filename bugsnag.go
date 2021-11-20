package spcontext

import (
	"runtime"
	"strings"

	bugsnagerrors "github.com/bugsnag/bugsnag-go/errors"
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

func (e *errorWithStackFrames) Error() string {
	return e.err.Error()
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
