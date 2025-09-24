package typex

import (
	"testing"
)

func TestSliceEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected bool
	}{
		{
			name:     "equal slices",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "empty slices",
			a:        []int{},
			b:        []int{},
			expected: true,
		},
		{
			name:     "nil slices",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "different lengths",
			a:        []int{1, 2, 3},
			b:        []int{1, 2},
			expected: false,
		},
		{
			name:     "different elements",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 4},
			expected: false,
		},
		{
			name:     "one nil, one empty",
			a:        nil,
			b:        []int{},
			expected: true,
		},
		{
			name:     "one nil, one non-empty",
			a:        nil,
			b:        []int{1},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SliceEqual(test.a, test.b)
			if result != test.expected {
				t.Errorf("SliceEqual(%v, %v) = %v, want %v", test.a, test.b, result, test.expected)
			}
		})
	}
}

func TestSliceEqualFunc(t *testing.T) {
	// Comparison function to consider nil and empty strings as equal.
	eq := func(a, b *string) bool {
		isEmpty := func(s *string) bool {
			return s == nil || (s != nil && *s == "")
		}
		if isEmpty(a) && isEmpty(b) {
			return true
		}
		if isEmpty(a) || isEmpty(b) {
			return false
		}
		return *a == *b
	}

	// Just a convenience function to easily get a pointer to a literal string.
	p := func(s string) *string { return &s }

	tests := []struct {
		name     string
		a        []*string
		b        []*string
		expected bool
	}{
		{
			name:     "equal non-empty strings",
			a:        []*string{p("foo"), p("bar")},
			b:        []*string{p("foo"), p("bar")},
			expected: true,
		},
		{
			name:     "nil and empty string treated as equal",
			a:        []*string{nil, p("bar")},
			b:        []*string{p(""), p("bar")},
			expected: true,
		},
		{
			name:     "both nil",
			a:        []*string{nil, nil},
			b:        []*string{nil, nil},
			expected: true,
		},
		{
			name:     "empty slices",
			a:        []*string{},
			b:        []*string{},
			expected: true,
		},
		{
			name:     "one nil slice",
			a:        []*string{},
			b:        nil,
			expected: true,
		},
		{
			name:     "different lengths",
			a:        []*string{p("foo")},
			b:        []*string{p("foo"), p("bar")},
			expected: false,
		},
		{
			name:     "different values",
			a:        []*string{p("foo")},
			b:        []*string{p("bar")},
			expected: false,
		},
		{
			name:     "nil and non-empty string",
			a:        []*string{nil},
			b:        []*string{p("baz")},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SliceEqualFunc(test.a, test.b, eq)
			if result != test.expected {
				t.Errorf("SliceEqualFunc() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestPtrEqual(t *testing.T) {
	val1 := "a string"
	val2 := "a string"
	val3 := "another string"

	tests := []struct {
		name     string
		a        *string
		b        *string
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "first nil",
			a:        nil,
			b:        &val1,
			expected: false,
		},
		{
			name:     "second nil",
			a:        &val1,
			b:        nil,
			expected: false,
		},
		{
			name:     "equal values",
			a:        &val1,
			b:        &val2,
			expected: true,
		},
		{
			name:     "different values",
			a:        &val1,
			b:        &val3,
			expected: false,
		},
		{
			name:     "same pointer",
			a:        &val1,
			b:        &val1,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PtrEqual(test.a, test.b)
			if result != test.expected {
				t.Errorf("PtrEqual() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestPtrCompare(t *testing.T) {
	val1 := 10
	val2 := 20

	less := func(a, b int) bool { return a < b }
	greater := func(a, b int) bool { return a > b }

	tests := []struct {
		name     string
		a        *int
		b        *int
		compare  func(int, int) bool
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			compare:  less,
			expected: true,
		},
		{
			name:     "first nil",
			a:        nil,
			b:        &val1,
			compare:  less,
			expected: false,
		},
		{
			name:     "second nil",
			a:        &val1,
			b:        nil,
			compare:  less,
			expected: false,
		},
		{
			name:     "less comparison true",
			a:        &val1,
			b:        &val2,
			compare:  less,
			expected: true,
		},
		{
			name:     "less comparison false",
			a:        &val2,
			b:        &val1,
			compare:  less,
			expected: false,
		},
		{
			name:     "greater comparison true",
			a:        &val2,
			b:        &val1,
			compare:  greater,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PtrCompare(test.a, test.b, test.compare)
			if result != test.expected {
				t.Errorf("PtrCompare() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestRefCompare(t *testing.T) {
	// Test with slices (reference types).
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5, 6}
	slice3 := []int{7, 8, 9, 10}

	tests := []struct {
		name     string
		a        []int
		b        []int
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "first nil",
			a:        nil,
			b:        slice1,
			expected: false,
		},
		{
			name:     "second nil",
			a:        slice1,
			b:        nil,
			expected: false,
		},
		{
			name:     "equal slices",
			a:        slice1,
			b:        slice2,
			expected: true,
		},
		{
			name:     "different slices",
			a:        slice1,
			b:        slice3,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := RefCompare(test.a, test.b, func(a, b []int) bool {
				// Makes two slices equal if they have the same length.
				return len(a) == len(b)
			})
			if result != test.expected {
				t.Errorf("RefCompare() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected bool
	}{
		{
			name:     "a less than b",
			a:        1,
			b:        2,
			expected: true,
		},
		{
			name:     "a greater than b",
			a:        2,
			b:        1,
			expected: false,
		},
		{
			name:     "a equal to b",
			a:        1,
			b:        1,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Less(test.a, test.b)
			if result != test.expected {
				t.Errorf("Less(%d, %d) = %v, want %v", test.a, test.b, result, test.expected)
			}
		})
	}
}

func TestIsNil(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "nil interface",
			value:    nil,
			expected: true,
		},
		{
			name:     "nil int pointer",
			value:    (*int)(nil),
			expected: true,
		},
		{
			name:     "nil string slice",
			value:    ([]string)(nil),
			expected: true,
		},
		{
			name:     "nil map",
			value:    (map[string]int)(nil),
			expected: true,
		},
		{
			name:     "nil channel",
			value:    (chan int)(nil),
			expected: true,
		},
		{
			name:     "nil function",
			value:    (func())(nil),
			expected: true,
		},
		{
			name:     "non-nil slice",
			value:    []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "empty slice",
			value:    []int{},
			expected: false,
		},
		{
			name:     "non-nil map",
			value:    map[string]int{"key": 1},
			expected: false,
		},
		{
			name:     "string value",
			value:    "hello",
			expected: false,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsNil(test.value)
			if result != test.expected {
				t.Errorf("IsNil(%v) = %v, want %v", test.value, result, test.expected)
			}
		})
	}
}
