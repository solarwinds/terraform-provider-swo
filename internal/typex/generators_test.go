package typex

import (
	"testing"
)

func TestZero(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		result := Zero[int]()
		if result != 0 {
			t.Errorf("Zero[int]() = %v, want 0", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result := Zero[string]()
		if result != "" {
			t.Errorf("Zero[string]() = %q, want \"\"", result)
		}
	})

	t.Run("float64", func(t *testing.T) {
		result := Zero[float64]()
		if result != 0.0 {
			t.Errorf("Zero[float64]() = %v, want 0.0", result)
		}
	})

	t.Run("slice", func(t *testing.T) {
		result := Zero[[]int]()
		if result != nil {
			t.Errorf("Zero[[]int]() = %v, want nil", result)
		}
	})

	t.Run("map", func(t *testing.T) {
		result := Zero[map[string]int]()
		if result != nil {
			t.Errorf("Zero[map[string]int]() = %v, want nil", result)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		result := Zero[*int]()
		if result != nil {
			t.Errorf("Zero[*int]() = %v, want nil", result)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type testStruct struct {
			name string
			age  int
		}
		result := Zero[testStruct]()
		expected := testStruct{name: "", age: 0}
		if result != expected {
			t.Errorf("Zero[testStruct]() = %v, want %v", result, expected)
		}
	})

	t.Run("interface", func(t *testing.T) {
		result := Zero[interface{}]()
		if result != nil {
			t.Errorf("Zero[interface{}]() = %v, want nil", result)
		}
	})
}
