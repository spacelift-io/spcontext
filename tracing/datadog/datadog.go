package datadog

import (
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/spcontext/tracing/internal"
)

// Tracer is an Datadog implementation of a Tracer.
type Tracer struct {
}

// OnSpanStart is called when a new span is created.
func (t *Tracer) OnSpanStart(ctx *spcontext.Context, name, resource string) *spcontext.Context {
	opts := []tracer.StartSpanOption{tracer.Measured()}
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
		// App Analytics (https://docs.datadoghq.com/tracing/legacy_app_analytics) is deprecated.
		// After its Rust rewrite, the Datadog Lambda Extension stopped tagging spans with analytics_enabled:true.
		// The recommended way is now to use retention filters; for that, we set spcontext.analyze:true.
		span.SetTag(ext.AnalyticsEvent, true)
		span.SetTag("spcontext.analyze", true)
	}

	// Datadog seems to be OK with duplicate tags but when testing we still want
	// make sure that the right (latest) value prevails.
	for key, value := range internal.DeduplicateFields(fields) {
		span.SetTag(key, value)
	}

	span.Finish(tracer.WithError(internal.UnwrapError(err)))
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
