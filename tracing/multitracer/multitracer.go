package multitracer

import "github.com/spacelift-io/spcontext"

// multitracer is a Tracer which calls all the given tracers.
type multitracer []spcontext.Tracer

// New creates a new Tracer which will call all the given tracers.
func New(tracers ...spcontext.Tracer) spcontext.Tracer {
	return multitracer(tracers)
}

// OnSpanStart is called when a new span is created.
func (t multitracer) OnSpanStart(ctx *spcontext.Context, name, resource string) *spcontext.Context {
	for _, tracer := range t {
		ctx = tracer.OnSpanStart(ctx, name, resource)
	}

	return ctx
}

// OnSpanClose is called when a span is closed.
func (t multitracer) OnSpanClose(ctx *spcontext.Context, err error, fields []any, drop, analyze bool) {
	for _, tracer := range t {
		tracer.OnSpanClose(ctx, err, fields, drop, analyze)
	}
}

// GetLogFields returns the fields which should be used in a log message in this
// context.
func (t multitracer) GetLogFields(ctx *spcontext.Context) []any {
	var fields []any

	for _, tracer := range t {
		fields = append(fields, tracer.GetLogFields(ctx)...)
	}

	return fields
}
