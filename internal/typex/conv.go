package typex

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/constraints"
)

// Map returns a slice where each element results from applying the given function f
// to the corresponding element in the input slice.
func Map[T, R any](input []T, f func(T) R) []R {
	out := make([]R, len(input))
	for i, v := range input {
		out[i] = f(v)
	}
	return out
}

// MapWithError is like Map, but the mapping function can return an error. If any
// invocation of f returns an error, the mapping stops and this error is returned,
// alongside a nil result.
func MapWithError[T, R any](input []T, f func(T) (R, error)) ([]R, error) {
	var err error
	out := make([]R, len(input))
	for i, v := range input {
		out[i], err = f(v)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// SliceShallowClone returns a (shallow) copy of the given slice, using the fastest
// method possible: pre-allocating the slice and using 'copy'. This is faster than
// the usual append(nil, source...).
func SliceShallowClone[T any](source []T) []T {
	dest := make([]T, len(source))
	copy(dest, source)
	return dest
}

// DerefOrDefault returns the value pointed to by the given pointer, if the latter is
// not nil. Otherwise, it returns the given default. This is straightforward code to
// do in place, but it takes a few lines and may needlessly obscure the purpose.
func DerefOrDefault[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// CastIntPtr converts a pointer to an integer type into a pointer to another integer
// type, by dereferencing and copying the value when needed (i.e., when the given
// pointer is not nil). Note that loss of precision might occur, depending on the type
// you're converting to.
func CastIntPtr[T, U constraints.Integer](ptr *T) *U {
	if ptr == nil {
		return nil
	}
	v := U(*ptr)
	return &v
}

// StringSliceToList converts a slice of strings into a Terraform List type. Any
// failure in building the list is recorded into the given diagnostics. Note that
// an empty input slice will produce an empty (rather than null) list.
func StringSliceToList(input []string, diags *diag.Diagnostics) types.List {
	stringValues := Map(input, func(s string) attr.Value {
		return types.StringValue(s)
	})
	list, d := types.ListValue(types.StringType, stringValues)
	diags.Append(d...)
	return list
}

// StringPtrSliceToList converts a slice of string pointers into a Terraform List.
// Elements are expected to be non-nil; if a nil pointer is found, the method will add
// an error to diagnostics and return a null list.
func StringPtrSliceToList(input []*string, diags *diag.Diagnostics) types.List {
	// Linting is disabled here. The error is internal; it's never returned to the caller.
	errNilValue := errors.New("unexpected nil pointer") //nolint:err113
	inputStr, err := MapWithError(input, func(s *string) (string, error) {
		if s == nil {
			return "", errNilValue
		}
		return *s, nil
	})
	if err != nil {
		diags.AddError("Missing Value Found",
			"unexpected nil element found in string pointer slice")
		return types.ListNull(types.StringType)
	}
	return StringSliceToList(inputStr, diags)
}

// CoalesceBy takes an equivalence function and two values of any type T, namely a
// current value and a replacement, and returns the latter unless the two values are
// equivalent according to the given function. Otherwise, it returns the current value.
func CoalesceBy[T any](current T, replacement T, isEquivalent func(T, T) bool) T {
	if !isEquivalent(current, replacement) {
		return replacement
	}
	return current
}

type ModelCollection interface {
	Elements() []attr.Value
}

// CollectionCoalesce takes two Terraform collection types (List, Set) and returns
// the replacement value unless both collections are empty, in which case it returns
// the current value. This is meant to be used when returning attributes for the state,
// so that nil vs empty collections do not produce unnecessary diffs.
func CollectionCoalesce[T ModelCollection](current, replacement T) T {
	return CoalesceBy(current, replacement, func(a, b T) bool {
		return len(a.Elements()) == 0 && len(b.Elements()) == 0
	})
}

// StringCoalesce takes two Terraform String types and returns the replacement value
// unless both strings are empty, in which case it returns the current value. Just like
// CollectionCoalesce, this is meant to be used when returning attributes for the state,
// to avoid unnecessary diffs caused by empty vs nil types.String.
func StringCoalesce(current, replacement types.String) types.String {
	return CoalesceBy(current, replacement, func(a, b types.String) bool {
		return len(a.ValueString()) == 0 && len(b.ValueString()) == 0
	})
}

// ObjectAttributesCoalesce takes two Terraform Object types and returns a new Object,
// where each attribute is coalesced according to its type. Currently, only String and
// collection (List, Set) attributes are coalesced. Other attributes are just copied
// from the replacement object.
func ObjectAttributesCoalesce(current, replacement types.Object) types.Object {
	currentMap := current.Attributes()
	result := make(map[string]attr.Value)

	for key, rv := range replacement.Attributes() {
		out := rv
		switch rv.(type) {
		case types.String:
			if c, ok := currentMap[key].(types.String); ok {
				out = StringCoalesce(c, rv.(types.String))
			} else {
				out = rv
			}
		case types.List:
			if c, ok := currentMap[key].(types.List); ok {
				out = CollectionCoalesce(c, rv.(types.List))
			} else {
				out = rv
			}
		case types.Set:
			if c, ok := currentMap[key].(types.Set); ok {
				out = CollectionCoalesce(c, rv.(types.Set))
			} else {
				out = rv
			}
		}
		result[key] = out
	}

	return types.ObjectValueMust(replacement.AttributeTypes(context.Background()), result)
}
