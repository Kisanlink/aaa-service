package test

import (
	"testing"
)

// TestExample is a simple test to verify the testing setup
func TestExample(t *testing.T) {
	t.Log("Testing setup is working correctly")

	// Test that basic assertions work
	expected := "hello"
	actual := "hello"

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestMath verifies basic math operations for testing framework
func TestMath(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"addition", 2, 3, 5},
		{"zero", 0, 5, 5},
		{"negative", -1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a + tt.b
			if result != tt.want {
				t.Errorf("got %d, want %d", result, tt.want)
			}
		})
	}
}

// BenchmarkExample provides a simple benchmark test
func BenchmarkExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "hello" + "world"
	}
}
