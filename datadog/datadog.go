package datadog

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/spacelift-io/spcontext"
)

// Tracer is an Datadog implementation of a Tracer.
type Tracer struct {
}

// OnSpanStart is called when a new span is created.
func (t *Tracer) OnSpanStart(ctx *spcontext.Context, name, resource string) *spcontext.Context {
	opts := []ddtrace.StartSpanOption{tracer.Measured()}
	if resource != "" {
		opts = append(opts, tracer.ResourceName(resource))
	}

	_, newCtx := tracer.StartSpanFromContext(ctx, name, opts...)
	return spcontext.FromStdContext(newCtx)
}

// OnSpanClose is called when a span is closed.
func (t *Tracer) OnSpanClose(ctx *spcontext.Context, err error, fields []interface{}, drop, analyze bool) {
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		ctx.Warnf("No span in context.")
		return
	}

	if drop {
		span.SetTag(ext.ManualDrop, true)
	}

	if analyze || (err != nil && !drop) {
		span.SetTag(ext.AnalyticsEvent, true)
	}

	addFields(span, fields...)
	span.Finish(tracer.WithError(err))
}

func addFields(span tracer.Span, fields ...interface{}) {
	// Datadog seems to be OK with duplicate tags but when testing we still want
	// make sure that the right (latest) value prevails.
	deduplicated := make(map[string]interface{})

	for i := 0; i < len(fields)/2; i++ {
		key := fields[2*i].(string)
		value := fields[2*i+1]
		deduplicated[key] = value
	}

	for key, value := range deduplicated {
		span.SetTag(key, value)
	}
}

// GetLogFields returns the fields which should be used in a log message in this context.
func (t *Tracer) GetLogFields(ctx *spcontext.Context) []interface{} {
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return nil
	}
	return []interface{}{
		"dd.trace_id", span.Context().TraceID(),
		"dd.span_id", span.Context().SpanID(),
	}
}
