package embedx

import "testing"

func TestNewSearchConfigDefaults(t *testing.T) {
	cfg := NewSearchConfig()
	if cfg.K != 10 {
		t.Errorf("Expected default K=10, got %d", cfg.K)
	}
	if cfg.Filter != nil {
		t.Errorf("Expected nil default Filter, got %v", cfg.Filter)
	}
}

func TestNewSearchConfigOptions(t *testing.T) {
	f := Eq("type", "doc")
	cfg := NewSearchConfig(WithK(3), WithFilter(f))
	if cfg.K != 3 {
		t.Errorf("Expected K=3, got %d", cfg.K)
	}
	if cfg.Filter == nil {
		t.Fatal("Expected Filter to be set")
	}
	if !cfg.Filter(map[string]any{"type": "doc"}) {
		t.Error("Filter should match type=doc")
	}
}

func TestEq(t *testing.T) {
	f := Eq("type", "doc")
	if !f(map[string]any{"type": "doc"}) {
		t.Error("Eq should match equal value")
	}
	if f(map[string]any{"type": "entity"}) {
		t.Error("Eq should not match different value")
	}
	if f(map[string]any{}) {
		t.Error("Eq should not match missing key")
	}
	if f(nil) {
		t.Error("Eq should not match nil metadata")
	}
	if !f(map[string]any{"type": "doc", "extra": 1}) {
		t.Error("Eq should match when key present among others")
	}
}

func TestIn(t *testing.T) {
	f := In("lang", "go", "rust")
	if !f(map[string]any{"lang": "go"}) {
		t.Error("In should match a present value")
	}
	if !f(map[string]any{"lang": "rust"}) {
		t.Error("In should match another present value")
	}
	if f(map[string]any{"lang": "python"}) {
		t.Error("In should not match absent value")
	}
	if f(map[string]any{}) {
		t.Error("In should not match missing key")
	}
	if f(nil) {
		t.Error("In should not match nil metadata")
	}
	if In("lang")(map[string]any{"lang": "go"}) {
		t.Error("In with no values should never match")
	}
}

func TestExists(t *testing.T) {
	f := Exists("payload_hash")
	if !f(map[string]any{"payload_hash": "abc"}) {
		t.Error("Exists should match present key")
	}
	if f(map[string]any{}) {
		t.Error("Exists should not match missing key")
	}
	if f(nil) {
		t.Error("Exists should not match nil metadata")
	}
}

func TestAndOrNot(t *testing.T) {
	positive := Eq("type", "doc")
	hasHash := Exists("payload_hash")

	and := And(positive, hasHash)
	if !and(map[string]any{"type": "doc", "payload_hash": "x"}) {
		t.Error("And should match when all match")
	}
	if and(map[string]any{"type": "doc"}) {
		t.Error("And should not match when one fails")
	}

	or := Or(positive, hasHash)
	if !or(map[string]any{"type": "doc"}) {
		t.Error("Or should match when any matches")
	}
	if !or(map[string]any{"payload_hash": "x"}) {
		t.Error("Or should match when the other matches")
	}
	if or(map[string]any{}) {
		t.Error("Or should not match when none match")
	}

	not := Not(positive)
	if !not(map[string]any{"type": "entity"}) {
		t.Error("Not should invert match")
	}
	if not(map[string]any{"type": "doc"}) {
		t.Error("Not should fail when inner matches")
	}

	if !And()(map[string]any{}) {
		t.Error("And with no filters should always match")
	}
	if Or()(map[string]any{}) {
		t.Error("Or with no filters should never match")
	}
}
