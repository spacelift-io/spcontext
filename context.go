package spcontext

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/go-kit/log"
	pkgerrors "github.com/pkg/errors"
)

// FieldsTab is the tab in bugsnag to put metadata fields into.
const FieldsTab = "fields"

// Valuer can be passed to get dynamic values in log fields.
type Valuer log.Valuer

// Notifier models the bugsnag interface.
type Notifier interface {
	Notify(error, ...interface{}) error
	AutoNotify(...interface{})
}

// Logger models the accepted logger underlying the context.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// Context is a drop-in replacement to context.Context. Include logging, error reporting and structured metadata.
type Context struct {
	context.Context
	fields   *Fields
	logger   Logger
	Notifier Notifier

	Tracer Tracer

	onSpanStartHooks []func(Span, Span)
}

// ContextOption is used to optionally configure the context on creation.
type ContextOption func(ctx *Context)

// WithNotifier adds an optional notifier to the new context.
func WithNotifier(notifier Notifier) ContextOption {
	return func(ctx *Context) {
		ctx.Notifier = notifier
	}
}

// WithTracer adds an optional tracer to the new context.
func WithTracer(tracer Tracer) ContextOption {
	return func(ctx *Context) {
		ctx.Tracer = tracer
	}
}

// OnSpanStart adds an on span start hook to the new context.
func OnSpanStart(hook func(parentSpan, activeSpan Span)) ContextOption {
	return func(ctx *Context) {
		ctx.onSpanStartHooks = append(ctx.onSpanStartHooks, hook)
	}
}

// New creates a new context with the logger and configured using additional options.
func New(logger Logger, opts ...ContextOption) *Context {
	ctx := &Context{
		Context: context.Background(),
		fields: (&Fields{}).With(
			"caller", Valuer(log.Caller(4)),
			"ts", Valuer(log.Timestamp(time.Now)),
		),
		logger: logger,
		Tracer: &NopTracer{},
	}

	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// With adds the given alternating keys and values to the context, returning a new child context.
func (ctx *Context) With(kvs ...interface{}) *Context {
	return &Context{
		Context:          ctx.Context,
		fields:           ctx.fields.With(kvs...),
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}
}

type contextKey struct{}

// Value returns the value for the given key in this context stack.
func (ctx *Context) Value(key interface{}) interface{} {
	if _, ok := key.(contextKey); ok {
		return ctx
	}
	return ctx.Context.Value(key)
}

// FromStdContext tries to find a spcontext.Context inside the given context.Context and returns a new one based on it.
// If no spcontext.Context is found, a default noop Context is returned.
func FromStdContext(stdCtx context.Context) *Context {
	v := stdCtx.Value(contextKey{})
	if v != nil {
		outCtx := v.(*Context)
		return &Context{
			Context:          stdCtx,
			fields:           outCtx.fields,
			logger:           outCtx.logger,
			Notifier:         outCtx.Notifier,
			Tracer:           outCtx.Tracer,
			onSpanStartHooks: outCtx.onSpanStartHooks,
		}
	}

	return &Context{
		Context:  stdCtx,
		fields:   &Fields{},
		logger:   log.NewNopLogger(),
		Notifier: nil,
		Tracer:   &NopTracer{},
	}
}

// WithValue adds a key value to the context. Use instead of context.WithValue
func WithValue(ctx *Context, key, val interface{}) *Context {
	return &Context{
		Context:          context.WithValue(ctx.Context, key, val),
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}
}

// CancelFunc is a function you can call to cancel the connected context.
type CancelFunc = context.CancelFunc

// WithCancel returns a cancelable context. Use instead of context.WithCancel
func WithCancel(ctx *Context) (*Context, CancelFunc) {
	newCtx, cancel := context.WithCancel(ctx.Context)
	return &Context{
		Context:          newCtx,
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}, cancel
}

// WithTimeout returns a context with a timeout. Use instead of context.WithTimeout.
func WithTimeout(ctx *Context, timeout time.Duration) (*Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx.Context, timeout)
	return &Context{
		Context:          newCtx,
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}, cancel
}

// WithTimeoutCause returns a context with a timeout,
// but also sets the cause of the returned Context when the timeout expires.
// The returned [CancelFunc] does not set the cause. Use instead of context.WithTimeoutCause.
func WithTimeoutCause(ctx *Context, timeout time.Duration, cause error) (*Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeoutCause(ctx.Context, timeout, cause)
	return &Context{
		Context:          newCtx,
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}, cancel
}

// WithDeadline returns a context with a deadline. Use instead of context.WithDeadline.
func WithDeadline(ctx *Context, d time.Time) (*Context, context.CancelFunc) {
	newCtx, cancel := context.WithDeadline(ctx.Context, d)
	return &Context{
		Context:          newCtx,
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}, cancel
}

// WithDeadlineCause returns a context with a deadline,
// but also sets the cause of the returned Context when the deadline is exceeded.
// The returned [CancelFunc] does not set the cause. Use instead of context.WithDeadlineCause.
func WithDeadlineCause(ctx *Context, d time.Time, cause error) (*Context, context.CancelFunc) {
	newCtx, cancel := context.WithDeadlineCause(ctx.Context, d, cause)
	return &Context{
		Context:          newCtx,
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}, cancel
}

// BackgroundFrom creates a new context.Background() from the given Context.
// This keeps all metadata fields and the logger/notifier configuration.
// Use instead of context.Background().
func BackgroundFrom(ctx *Context) *Context {
	return &Context{
		Context:          context.Background(),
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}
}

// BackgroundWithValuesFrom creates a new background context from the given Context.
// This keeps all key-values, metadata fields and the logger/notifier configuration.
// Use instead of BackgroundFrom when you want to keep key-value information.
func BackgroundWithValuesFrom(ctx *Context) *Context {
	return &Context{
		Context:          &backgroundWithValuesContext{ctx: ctx},
		fields:           ctx.fields,
		logger:           ctx.logger,
		Notifier:         ctx.Notifier,
		Tracer:           ctx.Tracer,
		onSpanStartHooks: ctx.onSpanStartHooks,
	}
}

type backgroundWithValuesContext struct {
	ctx context.Context
}

func (f *backgroundWithValuesContext) Deadline() (deadline time.Time, ok bool) { return }
func (f *backgroundWithValuesContext) Done() <-chan struct{}                   { return nil }
func (f *backgroundWithValuesContext) Err() error                              { return nil }
func (f *backgroundWithValuesContext) Value(key interface{}) interface{}       { return f.ctx.Value(key) }

func (ctx *Context) getEvaluatedFields() []interface{} {
	return append(ctx.fields.EvaluateFields(), ctx.Tracer.GetLogFields(ctx)...)
}

func (ctx *Context) log(fields []interface{}, level string, format string, args ...interface{}) {
	fields = append(fields,
		"level", level,
		"msg", fmt.Sprintf(format, args...))
	_ = ctx.logger.Log(fields...)
}

// Errorf logs the message with error level.
func (ctx *Context) Errorf(format string, args ...interface{}) {
	ctx.log(ctx.getEvaluatedFields(), "error", format, args...)
}

// Warnf logs the message with warning level.
func (ctx *Context) Warnf(format string, args ...interface{}) {
	ctx.log(ctx.getEvaluatedFields(), "warning", format, args...)
}

// Infof logs the message with info level.
func (ctx *Context) Infof(format string, args ...interface{}) {
	ctx.log(ctx.getEvaluatedFields(), "info", format, args...)
}

// Debugf logs the message with debug level.
func (ctx *Context) Debugf(format string, args ...interface{}) {
	ctx.log(ctx.getEvaluatedFields(), "debug", format, args...)
}

// InternalMessage is an internal error message not meant for users.
type InternalMessage error

// SafeMessage is a safe, user-friendly error message.
type SafeMessage error

// DirectError directly notifies about the error, without caring which error is
// user-facing, and which isn't.
func (ctx *Context) DirectError(err error, message string) error {
	return ctx.error(ctx.getEvaluatedFields(), err, InternalMessage(errors.New(message)), SafeMessage(errors.New(message)))
}

// Error reports the error to the logger and Bugsnag, while returning an error
// with a user-safe message.
func (ctx *Context) Error(err error, internal InternalMessage, safe SafeMessage) error {
	return ctx.error(ctx.getEvaluatedFields(), err, internal, safe)
}

// InternalError reports an error with a generic user-facing message.
func (ctx *Context) InternalError(err error, message string) error {
	return ctx.error(ctx.getEvaluatedFields(), err, InternalMessage(errors.New(message)), SafeMessage(errors.New("internal error")))
}

// RawError reports an error wrapped in a message.
func (ctx *Context) RawError(err error, message string) error {
	wrapped := pkgerrors.Wrap(err, message)
	return ctx.error(ctx.getEvaluatedFields(), err, wrapped, wrapped)
}

// notifiedError is an error which has already been sent to bugsnag.
type notifiedError struct {
	error
}

func (ctx *Context) error(fields []interface{}, err error, internal InternalMessage, safe SafeMessage) error {
	if err == nil {
		return nil
	}
	if notifiedErr := (notifiedError{}); errors.As(err, &notifiedErr) {
		// This error has already been notified to bugsnag before.
		return safe
	}

	fieldsMap := make(map[string]interface{})
	for i := 0; i < len(fields)/2; i++ {
		fieldsMap[fields[2*i].(string)] = fields[2*i+1]
	}

	if ctx.Notifier != nil && !strings.Contains(err.Error(), context.Canceled.Error()) {
		if notifierError := ctx.Notifier.Notify(err, bugsnag.MetaData{FieldsTab: fieldsMap}, ctx); notifierError != nil {
			ctx.Errorf("error notifying the exception tracker: %v", notifierError)
		}
	}

	ctx.log(fields, "error", "%s: %v", internal.Error(), err)

	return notifiedError{error: safe}
}

// Fields returns the context fields.
func (ctx *Context) Fields() *Fields {
	return ctx.fields
}

// EvaluateBugsnagMetadata returns Bugsnag metadata with the evaluated fields.
func (ctx *Context) EvaluateBugsnagMetadata() bugsnag.MetaData {
	fields := ctx.getEvaluatedFields()

	fieldsMap := make(map[string]interface{})
	for i := 0; i < len(fields)/2; i++ {
		fieldsMap[fields[2*i].(string)] = fields[2*i+1]
	}

	return bugsnag.MetaData{FieldsTab: fieldsMap}
}
