package spcontext

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
)

// SpanConfig configures Span creation.
type SpanConfig struct {
	Tags      *Fields
	Operation string
	Resource  string
}

// SpanOption is used to modify the SpanConfig.
type SpanOption func(*SpanConfig)

// WithTags adds the tags to the Span.
func WithTags(tags ...interface{}) SpanOption {
	return func(cfg *SpanConfig) {
		cfg.Tags = cfg.Tags.With(tags...)
	}
}

// WithOperation sets an operation name on the Span.
// The name will be trimmed to allowed characters only.
func WithOperation(operation string, a ...interface{}) SpanOption {
	operation = fmt.Sprintf(operation, a...)

	return func(cfg *SpanConfig) {
		builder := strings.Builder{}
		for _, r := range operation {
			if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '.' {
				builder.WriteRune(r)
			}
		}
		cfg.Operation = builder.String()
	}
}

// WithResource sets the resource name on the span.
func WithResource(resource string, a ...interface{}) SpanOption {
	resource = fmt.Sprintf(resource, a...)

	return func(cfg *SpanConfig) {
		cfg.Resource = resource
	}
}

// SpanCloseConfig configures Span finalization.
type SpanCloseConfig struct {
	Drop, Analyze bool
}

// SpanCloseOption is used to modify the SpanCloseConfig.
type SpanCloseOption func(*SpanCloseConfig)

// WithAnalyze sets the Span to be analyzed.
func WithAnalyze() SpanCloseOption {
	return func(cfg *SpanCloseConfig) {
		cfg.Analyze = true
	}
}

// WithDrop sets the Span to be dropped.
func WithDrop() SpanCloseOption {
	return func(cfg *SpanCloseConfig) {
		cfg.Drop = true
	}
}

// Tracer is used to create spans.
type Tracer interface {
	OnSpanStart(ctx *Context, name, resource string) *Context
	OnSpanClose(ctx *Context, err error, fields []interface{}, drop, analyze bool)
	GetLogFields(ctx *Context) []interface{}
}

// Span is a single tracing span, which can be closed with the given error.
type Span interface {
	Analyze()
	Close(err error, opts ...SpanCloseOption)
	Drop()
	SetTags(tags ...interface{})
}

type activeSpanContextKey struct{}

// StartSpan starts a new span using the context fields as metadata.
// It returns a new context with attached trace and span IDs as metadata.
func (ctx *Context) StartSpan(opts ...SpanOption) (*Context, Span) {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndex(funcName, "/"); i != -1 {
		funcName = funcName[i:]
	}

	cfg := SpanConfig{
		Tags: &Fields{},
	}
	WithOperation(funcName)(&cfg)
	for _, opt := range opts {
		opt(&cfg)
	}

	newCtx := ctx.Tracer.OnSpanStart(ctx, cfg.Operation, cfg.Resource)
	activeSpan := &span{ctx: newCtx, fields: cfg.Tags}

	ctx.onStartSpan(activeSpan)

	return WithValue(newCtx, activeSpanContextKey{}, activeSpan), activeSpan
}

func (ctx *Context) onStartSpan(activeSpan *span) {
	if len(ctx.onSpanStartHooks) == 0 {
		return
	}
	parentSpan, ok := ctx.Value(activeSpanContextKey{}).(*span)
	if !ok {
		return
	}
	for _, fn := range ctx.onSpanStartHooks {
		fn(parentSpan, activeSpan)
	}
}

func (ctx *Context) ActiveSpan() interface{} {
	s, ok := ctx.Value(activeSpanContextKey{}).(*span)
	if !ok {
		return nil
	}
	return s
}

type span struct {
	ctx           *Context
	fields        *Fields
	analyze, drop bool
}

func (s *span) Analyze() {
	s.analyze = true
}

func (s *span) Close(err error, opts ...SpanCloseOption) {
	cfg := SpanCloseConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	fields := s.fields.EvaluateFields()

	s.ctx.Tracer.OnSpanClose(s.ctx, err, fields, s.drop || cfg.Drop, s.analyze || cfg.Analyze)
}

func (s *span) Drop() {
	s.drop = true
}

func (s *span) SetTags(tags ...interface{}) {
	s.fields = s.fields.With(tags...)
}

func (s *span) Value(key string) interface{} {
	return s.fields.Value(key)
}

// NopTracer is a Tracer which does nothing.
type NopTracer struct {
}

// OnSpanStart does nothing.
func (n *NopTracer) OnSpanStart(ctx *Context, name, resource string) *Context {
	return ctx
}

// OnSpanClose does nothing.
func (n *NopTracer) OnSpanClose(ctx *Context, err error, fields []interface{}, drop, analyze bool) {
}

// GetLogFields does nothing.
func (n *NopTracer) GetLogFields(ctx *Context) []interface{} {
	return nil
}
