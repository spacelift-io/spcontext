package opentelemetry

import (
	"fmt"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/spcontext/tracing/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is an OpenTelemetry implementation of a Tracer.
type Tracer struct {
}

// OnSpanStart is called when a new span is created.
func (t *Tracer) OnSpanStart(ctx *spcontext.Context, name, resource string) *spcontext.Context {
	var opts []trace.SpanStartOption
	if resource != "" {
		opts = append(opts, trace.WithAttributes(attribute.String("resource", resource)))
	}
	if name != "" {
		opts = append(opts, trace.WithAttributes(attribute.String("operation.name", name)))
	}

	parentContext := ctx

	existingParent := trace.SpanFromContext(ctx)
	if existingParent != nil && existingParent.SpanContext().IsValid() {
		parentContext = spcontext.FromStdContext(trace.ContextWithSpan(ctx, existingParent))
	}

	newCtx, _ := otel.GetTracerProvider().Tracer("spacelift.io/tracing").Start(parentContext, name, opts...)

	return spcontext.FromStdContext(newCtx)
}

// OnSpanClose is called when a span is closed.
func (t *Tracer) OnSpanClose(ctx *spcontext.Context, err error, fields []any, drop, analyze bool) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		ctx.Warnf("No span in context.")
		return
	}

	// Currently we don't have a way to drop a span in OpenTelemetry.
	// The main issue is that even if we drop a parent span, the child spans will still be recorded.

	for key, value := range internal.DeduplicateFields(fields) {
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}

	if err != nil {
		span.RecordError(internal.UnwrapError(err))
		span.SetStatus(codes.Error, "")
	}

	span.End(trace.WithStackTrace(err != nil))
}

// GetLogFields returns the fields which should be used in a log message in this context.
func (t *Tracer) GetLogFields(ctx *spcontext.Context) []any {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return nil
	}

	spanCtx := span.SpanContext()

	return []interface{}{
		"otel.trace_id", spanCtx.TraceID(),
		"otel.span_id", spanCtx.SpanID(),
	}
}
