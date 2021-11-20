package spcontext

// IgnoreLogError wraps a logging function that returns an error and provides one
// that does not.
type IgnoreLogError func(...interface{}) error

// Log suppresses the error from the underlying logging function.
func (i IgnoreLogError) Log(args ...interface{}) {
	i(args...)
}
