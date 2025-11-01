package embedx

import "context"

// private key type to avoid collisions
type engineKey struct{}

// WithEngine returns a new context that carries the provided Embedder.
func WithEngine(ctx context.Context, e *Embedder) context.Context {
	return context.WithValue(ctx, engineKey{}, e)
}

// EngineFromContext retrieves the Embedder stored in ctx.
// Returns nil if no engine is present or type assertion fails.
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

// FromContext is an alias for EngineFromContext
func FromContext(ctx context.Context) *Embedder {
	return EngineFromContext(ctx)
}
