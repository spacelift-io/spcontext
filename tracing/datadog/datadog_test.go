package datadog_test

import (
	"testing"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	"github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/spcontext/tracing/datadog"
)

func TestOnSpanCloseAnalyze(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		opts        []spcontext.SpanCloseOption
		wantAnalyze bool
	}{
		{
			name:        "span closed with WithAnalyze is analyzed",
			opts:        []spcontext.SpanCloseOption{spcontext.WithAnalyze()},
			wantAnalyze: true,
		},
		{
			name:        "errored span is auto-analyzed",
			err:         errors.New("boom"),
			wantAnalyze: true,
		},
		{
			name:        "plain span is not analyzed",
			wantAnalyze: false,
		},
		{
			name:        "dropped errored span is not analyzed",
			err:         errors.New("boom"),
			opts:        []spcontext.SpanCloseOption{spcontext.WithDrop()},
			wantAnalyze: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mt := mocktracer.Start()
			defer mt.Stop()

			ctx := spcontext.New(log.NewNopLogger(), spcontext.WithTracer(&datadog.Tracer{}))
			_, span := ctx.StartSpan(spcontext.WithOperation("test.operation"))
			span.Close(tc.err, tc.opts...)

			finished := mt.FinishedSpans()
			require.Len(t, finished, 1)


			if tc.wantAnalyze {
				// Setting ext.AnalyticsEvent surfaces as the "_dd1.sr.eausr"
				// metric on the span - that is what ships on the wire.
				require.Equal(t, float64(1), finished[0].Tag("_dd1.sr.eausr"))
				// Boolean tags are stored as strings in span metadata.
				require.Equal(t, "true", finished[0].Tag("spcontext.analyze"))
			} else {
				require.Nil(t, finished[0].Tag("_dd1.sr.eausr"))
				require.Nil(t, finished[0].Tag("spcontext.analyze"))
			}
		})
	}
}
