// Package embedx provides utilities for working with contexts that carry Embedder instances.
package embedx

import "context"

// engineKey is a private type to avoid collisions when using context values.
type engineKey struct{}

// WithEngine returns a new context that carries the provided Embedder.
// This allows passing the Embedder through context to functions that need it.
func WithEngine(ctx context.Context, e *Embedder) context.Context {
	return context.WithValue(ctx, engineKey{}, e)
}

// EngineFromContext retrieves the Embedder stored in the context.
// Returns nil if no Embedder is present in the context or if type assertion fails.
func EngineFromContext(ctx context.Context) *Embedder {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(engineKey{}); v != nil {
		if e, ok := v.(*Embedder); ok {
			return e
		}
	}
	return nil
}

// FromContext is an alias for EngineFromContext.
// It retrieves the Embedder stored in the context.
func FromContext(ctx context.Context) *Embedder {
	return EngineFromContext(ctx)
}
