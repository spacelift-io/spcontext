package spcontext

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initSentinelErr is created at package init, just like real production
// sentinels; its stack trace is just runtime/init frames.
var initSentinelErr = errors.New("init sentinel")

func TestHasInitTimeStackOnly(t *testing.T) {
	t.Run("sentinel created at package init is classified as init-only",
		func(t *testing.T) {
			st, ok := initSentinelErr.(stackTracer)
			require.True(t, ok)
			assert.True(t, hasInitTimeStackOnly(st))
		},
	)

	t.Run("error created inside a function is not classified as init-only",
		func(t *testing.T) {
			inFnErr := errors.New("created in a function")
			st, ok := inFnErr.(stackTracer)
			require.True(t, ok)
			assert.False(t, hasInitTimeStackOnly(st))
		},
	)

	t.Run("wrap of an init-time sentinel exposes a non-init wrap stack",
		func(t *testing.T) {
			wrapped := errors.Wrap(initSentinelErr, "context")
			st, ok := wrapped.(stackTracer)
			require.True(t, ok)
			assert.False(t, hasInitTimeStackOnly(st))
		},
	)
}

func TestIsInitOrRuntimeFrame(t *testing.T) {
	tests := map[string]bool{
		"runtime.goexit":                           true,
		"runtime.main":                             true,
		"runtime.doInit1":                          true,
		"main.main":                                true,
		"github.com/spacelift-io/spcontext.init":   true,
		"github.com/spacelift-io/spcontext.init.0": true,
		"github.com/spacelift-io/backend/server/resolvers.(*runResolver).Logs":  false,
		"github.com/spacelift-io/backend/shared/runlogs.(*Manager).GetLogsPage": false,
		"main.run": false,
	}

	for name, expected := range tests {
		t.Run(name,
			func(t *testing.T) {
				require.Equal(t, expected, isInitOrRuntimeFrame(name))
			},
		)
	}
}
