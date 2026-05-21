package spcontext_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/franela/goblin"
	"github.com/go-kit/log"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/spcontext/testutils"
)

func TestContext(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Context", func() {
		var suite suite.Suite

		var logBuffer *bytes.Buffer
		var notifier *testutils.MockNotifier

		var sut *spcontext.Context

		g.BeforeEach(func() {
			suite.SetT(t)

			logBuffer = bytes.NewBuffer(nil)
			notifier = new(testutils.MockNotifier)

			sut = spcontext.New(log.NewLogfmtLogger(logBuffer)).With("xtest", true)
		})

		withNotifier := func(errorMessage string) {
			sut.Notifier = notifier

			notifier.On(
				"Notify",
				mock.MatchedBy(func(in interface{}) bool {
					err, ok := in.(error)
					suite.Require().True(ok)
					suite.Require().EqualError(err, errorMessage)
					return true
				}),
				mock.MatchedBy(func(fieldSlice []interface{}) bool {
					return fieldSlice[0].(bugsnag.MetaData)[spcontext.FieldsTab]["xtest"] == true
				}),
			).Return(nil)
		}

		notifierCalled := func() {
			suite.True(notifier.AssertCalled(
				t,
				"Notify",
				mock.AnythingOfType("*spcontext.errorWithStackFrames"),
				mock.AnythingOfType("[]interface {}"),
			))
		}

		notifierNotCalled := func() {
			suite.True(notifier.AssertNotCalled(t, "Notify"))
		}

		g.Describe("DirectError", func() {
			message := "message"
			problem := errors.New("bacon")
			var err error

			g.JustBeforeEach(func() {
				err = sut.DirectError(problem, message)
			})

			g.Describe("without notifier", func() {
				g.It("returns error", func() {
					suite.EqualError(err, message)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="message: bacon"`))
				})
			})

			g.Describe("with notifier", func() {
				g.BeforeEach(func() {
					withNotifier("bacon")
				})

				g.It("reports to the Bugsnag", func() {
					notifierCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, message)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="message: bacon"`))
				})
			})

			g.Describe("with context canceled", func() {
				g.BeforeEach(func() {
					problem = context.Canceled
				})

				g.It("not reports to the Bugsnag", func() {
					notifierNotCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, message)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="message: context canceled"`))
				})
			})
		})

		g.Describe("Error", func() {
			const internal = "internal"
			const safe = "safe"
			problem := errors.New("bacon")
			var err error

			g.JustBeforeEach(func() {
				err = sut.Error(problem, errors.New(internal), errors.New(safe))
			})

			g.Describe("without notifier", func() {
				g.It("returns error", func() {
					suite.EqualError(err, safe)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: bacon"`))
				})
			})

			g.Describe("with notifier", func() {
				g.BeforeEach(func() {
					withNotifier("bacon")
				})

				g.It("reports to the Bugsnag", func() {
					notifierCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, safe)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: bacon"`))
				})
			})

			g.Describe("with context canceled", func() {
				g.BeforeEach(func() {
					problem = context.Canceled
				})

				g.It("not reports to the Bugsnag", func() {
					notifierNotCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, safe)
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: context canceled"`))
				})
			})
		})

		g.Describe("InternalError", func() {
			const internal = "internal"
			problem := errors.New("bacon")
			var err error

			g.JustBeforeEach(func() {
				err = sut.InternalError(problem, internal)
			})

			g.Describe("without notifier", func() {
				g.It("returns error", func() {
					suite.EqualError(err, "internal error")
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: bacon"`))
				})
			})

			g.Describe("with notifier", func() {
				g.BeforeEach(func() {
					withNotifier("bacon")
				})

				g.It("reports to the Bugsnag", func() {
					notifierCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, "internal error")
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: bacon"`))
				})
			})

			g.Describe("with context canceled", func() {
				g.BeforeEach(func() {
					problem = context.Canceled
				})

				g.It("not reports to the Bugsnag", func() {
					notifierNotCalled()
				})

				g.It("returns error", func() {
					suite.EqualError(err, "internal error")
				})

				g.It("logs message", func() {
					Expect(logBuffer.String()).To(ContainSubstring(`level=error msg="internal: context canceled"`))
				})
			})
		})

		g.Describe("WithField", func() {
			var ret *spcontext.Context

			g.JustBeforeEach(func() {
				ret = sut.With("new", true)
			})

			g.Describe("without notifier", func() {
				g.It("returns spcontext", func() {
					suite.Require().NotNil(ret)
				})

				g.It("with fields", func() {
					fields := ret.Fields()
					suite.True(fields.Value("xtest").(bool))
					suite.True(fields.Value("new").(bool))

				})

				g.It("without notifier", func() {
					suite.Require().Nil(ret.Notifier)
				})
			})

			g.Describe("with notifier", func() {
				g.BeforeEach(func() {
					sut.Notifier = notifier
				})

				g.It("returns spcontext", func() {
					suite.Require().NotNil(ret)
				})

				g.It("with fields", func() {
					fields := ret.Fields()
					suite.True(fields.Value("xtest").(bool))
					suite.True(fields.Value("new").(bool))

				})

				g.It("with notifier", func() {
					suite.Require().Equal(notifier, ret.Notifier)
				})
			})
		})
	})
}

func TestBackgroundWithValuesFrom(t *testing.T) {
	base := spcontext.New(log.NewNopLogger())
	withField := base.With("fieldName", "fieldValue")
	withValue := spcontext.WithValue(withField, "keyName", "keyValue")
	withCancel, cancel := spcontext.WithCancel(withValue)
	backgroundWithValues := spcontext.BackgroundWithValuesFrom(withCancel)
	cancel()

	select {
	case <-withCancel.Done():
	default:
		t.Fatal("cancelled context should be done")
	}

	select {
	case <-backgroundWithValues.Done():
		t.Fatal("background context shouldn't be done")
	default:
	}

	require.Equal(t, "keyValue", backgroundWithValues.Value("keyName"))
	require.Equal(t, "fieldValue", backgroundWithValues.Fields().Value("fieldName"))
}

func TestNotifiedLogic(t *testing.T) {
	base := spcontext.New(log.NewNopLogger())
	notifier := new(testutils.MockNotifier)
	notifier.On(
		"Notify",
		mock.MatchedBy(func(in interface{}) bool {
			err, ok := in.(error)
			require.True(t, ok)
			assert.EqualError(t, err, "initial error")
			return true
		}),
		mock.Anything,
	).Times(1).Return(nil)
	defer notifier.AssertExpectations(t)
	base.Notifier = notifier
	err := base.InternalError(errors.New("initial error"), "foo")
	err = base.DirectError(err, "skipped error")
	_ = base.DirectError(err, "also skipped error")
}

func TestLogLevel(t *testing.T) {
	testCases := []struct {
		name        string
		logLevel    spcontext.LogLevel
		expectError bool
		expectWarn  bool
		expectInfo  bool
		expectDebug bool
	}{
		{
			name:        "Error level",
			logLevel:    spcontext.LogLevelError,
			expectError: true,
			expectWarn:  false,
			expectInfo:  false,
			expectDebug: false,
		},
		{
			name:        "Warn level",
			logLevel:    spcontext.LogLevelWarn,
			expectError: true,
			expectWarn:  true,
			expectInfo:  false,
			expectDebug: false,
		},
		{
			name:        "Info level",
			logLevel:    spcontext.LogLevelInfo,
			expectError: true,
			expectWarn:  true,
			expectInfo:  true,
			expectDebug: false,
		},
		{
			name:        "Debug level",
			logLevel:    spcontext.LogLevelDebug,
			expectError: true,
			expectWarn:  true,
			expectInfo:  true,
			expectDebug: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logBuffer := bytes.NewBuffer(nil)
			ctx := spcontext.New(log.NewLogfmtLogger(logBuffer), spcontext.WithLogLevel(tc.logLevel))

			// Test error logging
			ctx.Errorf("error message")
			if tc.expectError {
				assert.Contains(t, logBuffer.String(), `level=error msg="error message"`)
			} else {
				assert.NotContains(t, logBuffer.String(), `level=error msg="error message"`)
			}

			// Reset buffer and test warn logging
			logBuffer.Reset()
			ctx.Warnf("warn message")
			if tc.expectWarn {
				assert.Contains(t, logBuffer.String(), `level=warning msg="warn message"`)
			} else {
				assert.NotContains(t, logBuffer.String(), `level=warning msg="warn message"`)
			}

			// Reset buffer and test info logging
			logBuffer.Reset()
			ctx.Infof("info message")
			if tc.expectInfo {
				assert.Contains(t, logBuffer.String(), `level=info msg="info message"`)
			} else {
				assert.NotContains(t, logBuffer.String(), `level=info msg="info message"`)
			}

			// Reset buffer and test debug logging
			logBuffer.Reset()
			ctx.Debugf("debug message")
			if tc.expectDebug {
				assert.Contains(t, logBuffer.String(), `level=debug msg="debug message"`)
			} else {
				assert.NotContains(t, logBuffer.String(), `level=debug msg="debug message"`)
			}
		})
	}

	t.Run("Default log level is info", func(t *testing.T) {
		logBuffer := bytes.NewBuffer(nil)
		ctx := spcontext.New(log.NewLogfmtLogger(logBuffer))

		ctx.Debugf("debug message")
		ctx.Infof("info message")

		assert.NotContains(t, logBuffer.String(), `level=debug msg="debug message"`)
		assert.Contains(t, logBuffer.String(), `level=info msg="info message"`)
	})

	t.Run("Log level is preserved in child contexts", func(t *testing.T) {
		logBuffer := bytes.NewBuffer(nil)
		ctx := spcontext.New(log.NewLogfmtLogger(logBuffer), spcontext.WithLogLevel(spcontext.LogLevelWarn))
		
		childCtx := ctx.With("field", "value")
		childCtx.Infof("info message")
		childCtx.Warnf("warn message")

		assert.NotContains(t, logBuffer.String(), `level=info msg="info message"`)
		assert.Contains(t, logBuffer.String(), `level=warning msg="warn message"`)
	})
}
