package typex

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) string
		expected []string
	}{
		{
			name:     "convert ints to strings",
			input:    []int{1, 2, 3},
			fn:       func(i int) string { return strconv.Itoa(i) },
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "empty slice",
			input:    []int{},
			fn:       func(i int) string { return strconv.Itoa(i) },
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			fn:       func(i int) string { return strconv.Itoa(i) },
			expected: []string{},
		},
		{
			name:     "multiply by 2",
			input:    []int{1, 2, 3, 4},
			fn:       func(i int) string { return strconv.Itoa(i * 2) },
			expected: []string{"2", "4", "6", "8"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Map(test.input, test.fn)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Map() = %v, want %v", result, test.expected)
			}
		})
	}

}

func TestMapWithError(t *testing.T) {
	parseIntFunc := func(s string) (int, error) {
		return strconv.Atoi(s)
	}

	tests := []struct {
		name        string
		input       []string
		fn          func(string) (int, error)
		expected    []int
		expectError bool
	}{
		{
			name:        "valid integers",
			input:       []string{"1", "2", "3"},
			fn:          parseIntFunc,
			expected:    []int{1, 2, 3},
			expectError: false,
		},
		{
			name:        "empty slice",
			input:       []string{},
			fn:          parseIntFunc,
			expected:    []int{},
			expectError: false,
		},
		{
			name:        "invalid integer",
			input:       []string{"1", "invalid", "3"},
			fn:          parseIntFunc,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "first element invalid",
			input:       []string{"invalid", "2", "3"},
			fn:          parseIntFunc,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "last element invalid",
			input:       []string{"1", "2", "invalid"},
			fn:          parseIntFunc,
			expected:    nil,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := MapWithError(test.input, test.fn)

			if test.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, test.expected) {
					t.Errorf("MapWithError() = %v, want %v", result, test.expected)
				}
			}
		})
	}

	// Test custom error
	t.Run("custom error function", func(t *testing.T) {
		// Suppressing linter rule because it's okay to use dynamic errors here.
		customErr := errors.New("custom error") //nolint:err113
		failFunc := func(s string) (int, error) {
			if s == "fail" {
				return 0, customErr
			}
			return len(s), nil
		}

		result, err := MapWithError([]string{"hello", "fail", "world"}, failFunc)

		// Suppressing linter rule because we want to check for the exact error.
		if err != customErr { //nolint:err113
			t.Errorf("Expected custom error, got %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})
}

func TestSliceShallowClone(t *testing.T) {
	tests := []struct {
		name  string
		input []int
	}{
		{
			name:  "normal slice",
			input: []int{1, 2, 3, 4, 5},
		},
		{
			name:  "empty slice",
			input: []int{},
		},
		{
			name:  "single element",
			input: []int{42},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := make([]int, len(test.input))
			copy(original, test.input)

			clone := SliceShallowClone(test.input)

			// Check that clone has same content
			if !reflect.DeepEqual(clone, test.input) {
				t.Errorf("Clone = %v, want %v", clone, test.input)
			}

			if len(test.input) > 0 {
				// Check that they are different slices
				if &clone[0] == &test.input[0] {
					t.Error("Clone should not share underlying array")
				}

				test.input[0] = 999
				if !reflect.DeepEqual(clone, original) {
					t.Error("Clone was affected by modification of original")
				}
			}
		})
	}

	// Test with nil slice
	t.Run("nil slice", func(t *testing.T) {
		var nilSlice []int
		clone := SliceShallowClone(nilSlice)
		if clone == nil {
			t.Error("Expected non-nil clone of nil slice")
		}
		if len(clone) != 0 {
			t.Errorf("Expected empty clone, got %v", clone)
		}
	})
}

func TestDerefOrDefault(t *testing.T) {
	val1 := 42
	val2 := 100
	defaultVal := 999

	tests := []struct {
		name     string
		ptr      *int
		def      int
		expected int
	}{
		{
			name:     "non-nil pointer",
			ptr:      &val1,
			def:      defaultVal,
			expected: 42,
		},
		{
			name:     "nil pointer",
			ptr:      nil,
			def:      defaultVal,
			expected: 999,
		},
		{
			name:     "pointer to zero",
			ptr:      func() *int { i := 0; return &i }(),
			def:      defaultVal,
			expected: 0,
		},
		{
			name:     "different value",
			ptr:      &val2,
			def:      defaultVal,
			expected: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DerefOrDefault(test.ptr, test.def)
			if result != test.expected {
				t.Errorf("DerefOrDefault() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestCastIntPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "int32 to int64",
			input:    func() *int32 { i := int32(42); return &i }(),
			expected: func() *int64 { i := int64(42); return &i }(),
		},
		{
			name:     "int64 to int32",
			input:    func() *int64 { i := int64(100); return &i }(),
			expected: func() *int32 { i := int32(100); return &i }(),
		},
		{
			name:     "int to int8",
			input:    func() *int { i := 50; return &i }(),
			expected: func() *int8 { i := int8(50); return &i }(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch input := test.input.(type) {
			case *int32:
				result := CastIntPtr[int32, int64](input)
				expected := test.expected.(*int64)
				if result == nil || expected == nil {
					t.Fatalf("Unexpected nil values")
				}
				if *result != *expected {
					t.Errorf("CastIntPtr() = %v, want %v", *result, *expected)
				}
			case *int64:
				result := CastIntPtr[int64, int32](input)
				expected := test.expected.(*int32)
				if result == nil || expected == nil {
					t.Fatalf("Unexpected nil values")
				}
				if *result != *expected {
					t.Errorf("CastIntPtr() = %v, want %v", *result, *expected)
				}
			case *int:
				result := CastIntPtr[int, int8](input)
				expected := test.expected.(*int8)
				if result == nil || expected == nil {
					t.Fatalf("Unexpected nil values")
				}
				if *result != *expected {
					t.Errorf("CastIntPtr() = %v, want %v", *result, *expected)
				}
			}
		})
	}

	// Test nil pointer.
	t.Run("nil pointer", func(t *testing.T) {
		var nilPtr *int32
		result := CastIntPtr[int32, int64](nilPtr)
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %v", result)
		}
	})

	// Test conversion with loss of precision.
	t.Run("precision loss", func(t *testing.T) {
		large := int64(1000000)
		result := CastIntPtr[int64, int8](&large)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		// Suppressing linter rule because there is expected overflow here.
		if expected := int8(large); *result != expected { //nolint:gosec
			t.Errorf("Expected truncated value %v, got %v", expected, *result)
		}
	})
}

func TestStringSliceToList(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{
			name:  "normal string slice",
			input: []string{"hello", "world", "test"},
		},
		{
			name:  "empty slice",
			input: []string{},
		},
		{
			name:  "single element",
			input: []string{"single"},
		},
		{
			name:  "slice with empty strings",
			input: []string{"", "hello", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := StringSliceToList(test.input, &diags)

			if diags.HasError() {
				t.Fatalf("Unexpected error(s): %v", diags.Errors())
			}
			if result.IsNull() {
				t.Error("Expected non-null result")
			}

			elements := result.Elements()
			if len(elements) != len(test.input) {
				t.Errorf("Expected %d elements, got %d", len(test.input), len(elements))
			}

			for i, elem := range elements {
				if strVal, ok := elem.(types.String); ok {
					if strVal.ValueString() != test.input[i] {
						t.Errorf("Element %d: expected %q, got %q", i, test.input[i], strVal.ValueString())
					}
				} else {
					t.Errorf("Element %d is not a String type", i)
				}
			}
		})
	}

	// Test nil slice
	t.Run("nil slice", func(t *testing.T) {
		var diags diag.Diagnostics
		result := StringSliceToList(nil, &diags)

		if diags.HasError() {
			t.Fatalf("Unexpected error(s): %v", diags.Errors())
		}
		if result.IsNull() {
			t.Error("Expected non-null result for nil slice")
		}
		if len(result.Elements()) != 0 {
			t.Errorf("Expected empty list, got %d elements", len(result.Elements()))
		}
	})
}

func TestStringPtrSliceToList(t *testing.T) {
	str1 := "one"
	str2 := "two"
	str3 := "three"

	tests := []struct {
		name          string
		input         []*string
		expectSuccess bool
	}{
		{
			name:          "normal string pointer slice",
			input:         []*string{&str1, &str2, &str3},
			expectSuccess: true,
		},
		{
			name:          "empty slice",
			input:         []*string{},
			expectSuccess: true,
		},
		{
			name:          "single element",
			input:         []*string{&str1},
			expectSuccess: true,
		},
		{
			name:          "slice with nil pointer",
			input:         []*string{&str1, nil, &str3},
			expectSuccess: false,
		},
		{
			name:          "all nil pointers",
			input:         []*string{nil, nil},
			expectSuccess: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := StringPtrSliceToList(test.input, &diags)

			if test.expectSuccess {
				if diags.HasError() {
					t.Fatalf("Unexpected error(s): %v", diags.Errors())
				}
				if result.IsNull() {
					t.Errorf("Unexpected null result")
				}

				elements := result.Elements()
				if len(elements) != len(test.input) {
					t.Fatalf("Expected %d elements, got %d", len(test.input), len(elements))
				}
				for i, elem := range elements {
					if strVal, ok := elem.(types.String); ok {
						expected := *test.input[i]
						if strVal.ValueString() != expected {
							t.Errorf("Element %d: expected %q, got %q", i, expected, strVal.ValueString())
						}
					} else {
						t.Errorf("Element %d is not a String type", i)
					}
				}
			} else {
				if !diags.HasError() {
					t.Error("Expected error but got none")
				}
				if !result.IsNull() {
					t.Error("Expected null result on error")
				}
			}
		})
	}
}

func TestCoalesceBy(t *testing.T) {
	isEquivalent := func(a, b string) bool {
		return len(a) == len(b)
	}

	tests := []struct {
		name          string
		current       string
		replacement   string
		expectCurrent bool
	}{
		{
			name:          "equivalent values",
			current:       "current",
			replacement:   "replace",
			expectCurrent: true,
		},
		{
			name:          "different values",
			current:       "current",
			replacement:   "replacement",
			expectCurrent: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CoalesceBy(test.current, test.replacement, isEquivalent)
			expected := test.replacement
			if test.expectCurrent {
				expected = test.current
			}
			if result != expected {
				t.Errorf("CoalesceBy() = %s, want %s", result, expected)
			}
		})
	}
}

func TestCollectionCoalesce(t *testing.T) {
	nullList := types.ListNull(types.StringType)
	emptyList1 := types.ListValueMust(types.StringType, []attr.Value{})
	emptyList2 := types.ListValueMust(types.StringType, []attr.Value{})
	nonEmptyList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value")})

	tests := []struct {
		name          string
		current       types.List
		replacement   types.List
		expectCurrent bool
	}{
		{
			name:          "both empty",
			current:       emptyList1,
			replacement:   emptyList2,
			expectCurrent: true,
		},
		{
			name:          "current null, replacement empty",
			current:       nullList,
			replacement:   emptyList1,
			expectCurrent: true,
		},
		{
			name:          "current empty, replacement null",
			current:       emptyList1,
			replacement:   nullList,
			expectCurrent: true,
		},
		{
			name:          "current null, replacement non-empty",
			current:       nullList,
			replacement:   nonEmptyList,
			expectCurrent: false,
		},
		{
			name:          "current empty, replacement non-empty",
			current:       emptyList1,
			replacement:   nonEmptyList,
			expectCurrent: false,
		},
		{
			name:          "current non-empty, replacement empty",
			current:       nonEmptyList,
			replacement:   emptyList1,
			expectCurrent: false,
		},
		{
			name:          "both non-empty",
			current:       nonEmptyList,
			replacement:   nonEmptyList,
			expectCurrent: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CollectionCoalesce(test.current, test.replacement)
			expected := test.replacement
			if test.expectCurrent {
				expected = test.current
			}

			if !result.Equal(expected) {
				t.Errorf("Unexpected result: got %v, want %v", result, expected)
			}
		})
	}
}

func TestStringCoalesce(t *testing.T) {
	nullStr := types.StringNull()
	emptyStr := types.StringValue("")
	nonEmptyStr := types.StringValue("value")

	tests := []struct {
		name          string
		current       types.String
		replacement   types.String
		expectCurrent bool
	}{
		{
			name:          "both empty",
			current:       emptyStr,
			replacement:   nullStr,
			expectCurrent: true,
		},
		{
			name:          "current empty, replacement non-empty",
			current:       emptyStr,
			replacement:   nonEmptyStr,
			expectCurrent: false,
		},
		{
			name:          "current non-empty, replacement empty",
			current:       nonEmptyStr,
			replacement:   emptyStr,
			expectCurrent: false,
		},
		{
			name:          "both non-empty",
			current:       nonEmptyStr,
			replacement:   nonEmptyStr,
			expectCurrent: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringCoalesce(test.current, test.replacement)
			expected := test.replacement
			if test.expectCurrent {
				expected = test.current
			}

			if result.ValueString() != expected.ValueString() {
				t.Errorf("Expected %q, got %q", expected.ValueString(), result.ValueString())
			}
		})
	}
}

func TestObjectAttributesCoalesce(t *testing.T) {
	// Create test objects as maps.
	currentAttrs := map[string]attr.Value{
		"string_attr": types.StringNull(),
		"list_attr":   types.ListValueMust(types.StringType, []attr.Value{}),
		"set_attr":    types.SetValueMust(types.StringType, []attr.Value{}),
		"other_attr":  types.Int64Value(42),
	}

	replacementAttrs := map[string]attr.Value{
		"string_attr": types.StringValue(""),
		"list_attr":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value")}),
		"set_attr":    types.SetNull(types.StringType),
		"other_attr":  types.Int64Value(100),
	}

	attrTypes := map[string]attr.Type{
		"string_attr": types.StringType,
		"list_attr":   types.ListType{ElemType: types.StringType},
		"set_attr":    types.SetType{ElemType: types.StringType},
		"other_attr":  types.Int64Type,
	}

	// Create the actual Terraform objects and call the coalesce function.
	current := types.ObjectValueMust(attrTypes, currentAttrs)
	replacement := types.ObjectValueMust(attrTypes, replacementAttrs)
	result := ObjectAttributesCoalesce(current, replacement)
	resultAttrs := result.Attributes()

	// String attribute: both empty, should return current
	if strAttr, ok := resultAttrs["string_attr"].(types.String); ok {
		if !strAttr.Equal(currentAttrs["string_attr"]) {
			t.Error("Expected current value for string_attr")
		}
	} else {
		t.Error("string_attr should be a String type")
	}

	// List attribute: replacement non-empty, should return replacement
	if listAttr, ok := resultAttrs["list_attr"].(types.List); ok {
		if !listAttr.Equal(replacementAttrs["list_attr"]) {
			t.Error("Expected replacement list list_attr")
		}
	} else {
		t.Error("list_attr should be a List type")
	}

	// Set attribute: both empty, should return current
	if setAttr, ok := resultAttrs["set_attr"].(types.Set); ok {
		if !setAttr.Equal(currentAttrs["set_attr"]) {
			t.Error("Expected current value for set_attr")
		}
	} else {
		t.Error("set_attr should be a Set type")
	}

	// Other attribute: should be replaced
	if intAttr, ok := resultAttrs["other_attr"].(types.Int64); ok {
		if intAttr.ValueInt64() != 100 {
			t.Errorf("Expected other_attr to be 100, got %d", intAttr.ValueInt64())
		}
	} else {
		t.Error("other_attr should be Int64 type")
	}
}
