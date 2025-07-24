package spcontext

import (
	"context"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// ContextInjector injects the given context into each request.
// Swapping the underlying context.Context for the one in the request.
func ContextInjector(ctx *Context) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(w, r.WithContext(&Context{
			Context:          &mergeValuesContext{base: r.Context(), merged: ctx.Context},
			fields:           ctx.fields,
			logger:           ctx.logger,
			logLevel:         ctx.logLevel,
			Notifier:         ctx.Notifier,
			Tracer:           ctx.Tracer,
			onSpanStartHooks: ctx.onSpanStartHooks,
		}))
	}
}

// GRPCStreamContextInjector injects the given context into each stream.
// Swapping the underlying context.Context for the one in the request.
func GRPCStreamContextInjector(ctx *Context) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := &Context{
			Context:          &mergeValuesContext{base: stream.Context(), merged: ctx.Context},
			fields:           ctx.fields,
			logger:           ctx.logger,
			logLevel:         ctx.logLevel,
			Notifier:         ctx.Notifier,
			Tracer:           ctx.Tracer,
			onSpanStartHooks: ctx.onSpanStartHooks,
		}
		wrappedStream := grpc_middleware.WrapServerStream(stream)
		wrappedStream.WrappedContext = newCtx
		return handler(srv, wrappedStream)
	}
}

// mergeValuesContext merges values from two contexts, with other properties being based on the base one.
// Can be removed when this proposal gets implemented: https://github.com/golang/go/issues/36503
type mergeValuesContext struct {
	base, merged context.Context
}

func (m *mergeValuesContext) Deadline() (deadline time.Time, ok bool) {
	return m.base.Deadline()
}

func (m *mergeValuesContext) Done() <-chan struct{} {
	return m.base.Done()
}

func (m *mergeValuesContext) Err() error {
	return m.base.Err()
}

func (m *mergeValuesContext) Value(key interface{}) interface{} {
	if val := m.base.Value(key); val != nil {
		return val
	}
	return m.merged.Value(key)
}
