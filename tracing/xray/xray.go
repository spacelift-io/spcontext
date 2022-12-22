package xray

import (
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/spcontext/tracing/internal"
)

// Tracer is an AWS X-Ray implementation of a Tracer.
type Tracer struct {
}

// OnSpanStart is called when a new span is created.
func (t *Tracer) OnSpanStart(ctx *spcontext.Context, name, resource string) *spcontext.Context {
	createFn := xray.BeginSubsegment

	// Depending on whether the segment exists or not, we either create a
	// subsegment or a new segment.
	if xray.GetSegment(ctx) == nil {
		createFn = xray.BeginSegment
	}

	newCtx, segment := createFn(ctx, name)
	if resource != "" {
		if err := segment.AddAnnotation("resource", resource); err != nil {
			_ = ctx.DirectError(err, "failed to add resource annotation to an X-Ray segment")
		}
	}

	return spcontext.FromStdContext(newCtx)
}

// OnSpanClose is called when a span is closed.
func (t *Tracer) OnSpanClose(ctx *spcontext.Context, err error, fields []any, drop, analyze bool) {
	segment := xray.GetSegment(ctx)
	if segment == nil {
		ctx.Warnf("No segment in context.")
		return
	}

	setDropAndAnalyze(segment, drop, analyze)

	for key, value := range internal.DeduplicateFields(fields) {
		if err := segment.AddAnnotation(key, value); err != nil {
			_ = ctx.DirectError(err, "failed to add annotation to an X-Ray segment")
		}
	}

	segment.Close(err)
}

// GetLogFields returns the fields which should be used in a log message in this context.
func (t *Tracer) GetLogFields(ctx *spcontext.Context) []any {
	segment := xray.GetSegment(ctx)
	if segment == nil {
		return nil
	}

	segment.RLock()
	defer segment.RUnlock()

	return []any{
		"xray.trace_id", segment.TraceID,
		"xray.segment_id", segment.ID,
	}
}

// Not 100% sure about that. The docs aren't entirely clear on this and
// there are no references to these fields in the docs, but they are public so
// presumably they are meant to be used?
func setDropAndAnalyze(segment *xray.Segment, drop, analyze bool) {
	segment.Lock()
	defer segment.Unlock()

	if drop {
		segment.Dummy = true
	} else if analyze {
		segment.Sampled = true
	}
}
