package embedx

import (
	"context"
	"testing"
)

func TestContextFunctions(t *testing.T) {
	// Test WithEngine and EngineFromContext
	store := NewMemoryStore()
	engine := New(store)

	ctx := WithEngine(context.Background(), engine)

	// Test EngineFromContext
	retrievedEngine := EngineFromContext(ctx)
	if retrievedEngine != engine {
		t.Error("EngineFromContext did not retrieve the correct engine")
	}

	// Test FromContext (alias for EngineFromContext)
	retrievedEngine2 := FromContext(ctx)
	if retrievedEngine2 != engine {
		t.Error("FromContext did not retrieve the correct engine")
	}

	// Test with nil context
	nilEngine := EngineFromContext(nil)
	if nilEngine != nil {
		t.Error("EngineFromContext with nil context should return nil")
	}

	// Test with context that doesn't have engine
	emptyCtx := context.Background()
	emptyEngine := EngineFromContext(emptyCtx)
	if emptyEngine != nil {
		t.Error("EngineFromContext with empty context should return nil")
	}
}
