package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

// mockVectorStore implements VectorStore interface for testing
type mockVectorStore struct {
	data     map[string][]float32
	saveErr  error
	getErr   error
	allErr   error
	closeErr error
}

func (m *mockVectorStore) SaveVector(id string, vec []float32) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.data == nil {
		m.data = make(map[string][]float32)
	}
	m.data[id] = append([]float32(nil), vec...)
	return nil
}

func (m *mockVectorStore) GetVector(id string) ([]float32, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	vec, exists := m.data[id]
	if !exists {
		return nil, fmt.Errorf("vector not found: %s", id)
	}
	return vec, nil
}

func (m *mockVectorStore) GetAllVectors() (map[string][]float32, error) {
	if m.allErr != nil {
		return nil, m.allErr
	}
	return m.data, nil
}

func (m *mockVectorStore) Close() error {
	return m.closeErr
}

func TestCmdInit(t *testing.T) {
	cmd := cmdInit()

	// Test command properties
	if cmd.Use != "init" {
		t.Errorf("Expected Use to be 'init', got '%s'", cmd.Use)
	}
}

func TestCmdAdd(t *testing.T) {
	cmd := cmdAdd()

	if cmd.Use != "add [id] [v1 v2 v3 ...]" {
		t.Errorf("Expected Use to be 'add [id] [v1 v2 v3 ...]', got '%s'", cmd.Use)
	}
}

func TestCmdSearch(t *testing.T) {
	cmd := cmdSearch()

	if cmd.Use != "search [v1 v2 v3 ...]" {
		t.Errorf("Expected Use to be 'search [v1 v2 v3 ...]', got '%s'", cmd.Use)
	}
}

func TestParseFloat32Vec(t *testing.T) {
	// Test successful parsing
	vec, err := parseFloat32Vec([]string{"1.0", "2.5", "-3.7"})
	if err != nil {
		t.Errorf("parseFloat32Vec failed: %v", err)
	}

	expected := []float32{1.0, 2.5, -3.7}
	for i, expectedVal := range expected {
		if vec[i] != expectedVal {
			t.Errorf("Index %d: expected %f, got %f", i, expectedVal, vec[i])
		}
	}

	// Test invalid input
	_, err = parseFloat32Vec([]string{"invalid", "2.0"})
	if err == nil {
		t.Error("Expected error for invalid input, got nil")
	}

	// Test empty input
	vec, err = parseFloat32Vec([]string{})
	if err != nil {
		t.Errorf("Empty input should work, got error: %v", err)
	}
	if len(vec) != 0 {
		t.Error("Empty input should return empty vector")
	}
}

func TestCmdContext(t *testing.T) {
	// Test that commands properly retrieve engine from context
	mockStore := &mockVectorStore{}
	engine := embedx.New(mockStore)

	// Test EngineFromContext helper function (which is the actual exported function)
	ctx := embedx.WithEngine(context.Background(), engine)
	retrievedEngine := embedx.EngineFromContext(ctx)

	if retrievedEngine == nil {
		t.Error("EngineFromContext returned nil")
	}

	if retrievedEngine != engine {
		t.Error("EngineFromContext returned different engine")
	}

	// Test with nil context
	nilEngine := embedx.EngineFromContext(nil)
	if nilEngine != nil {
		t.Error("EngineFromContext with nil context should return nil")
	}
}
