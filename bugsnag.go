package spcontext

// BugsnagLogger wraps the given Context inside a bugsnag friendly logger.
type BugsnagLogger struct {
	Ctx Context
}

// Printf logs the message with info level.
func (l *BugsnagLogger) Printf(format string, v ...interface{}) {
	l.Ctx.Infof(format, v...)
}
