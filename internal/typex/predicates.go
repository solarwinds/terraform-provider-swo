package typex

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

// SliceEqual returns true if the two slices are exactly the same; i.e., they
// have the same elements in the same order. A usual practice in Go is resorting
// to reflect.DeepEquals. But that depends on reflection and is therefore slower.
// Note that this function is meant only for a shallow check, and it does consider
// an empty slice equal to a nil slice.
func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SliceEqualFunc returns true if the two given slices match. It works just
// like SliceEqual, except that it broadens the scope to any arbitrary type T
// by using a provided item equality function. Note that this function considers
// an empty slice equal to a nil slice.
func SliceEqualFunc[T any](a, b []T, eq func(T, T) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !eq(a[i], b[i]) {
			return false
		}
	}
	return true
}

// PtrEqual returns true when pointers to values of two comparable types ere
// either both nil, or both non-nil and they point to equal values.
func PtrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// LeftPtrEqual returns true when, given a pointer to a value of a comparable
// type T and a bare value of the same type, the former points to a value equal
// to the latter. In particular, it returns false when the pointer is nil.
func LeftPtrEqual[T comparable](a *T, b T) bool {
	if a == nil {
		return false
	}
	return *a == b
}

// PtrCompare returns true when pointers to values of any two types are either
// both nil, or both non-nil and the given comparison applied to the dereferenced
// values returns true. It works just like PtrEqual but with a generic comparison
// rather than equality.
func PtrCompare[T any](a, b *T, compare func(T, T) bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return compare(*a, *b)
}

// RefCompare returns true when two values of any reference type are either both
// nil, or both non-nil and the given comparison applied to them returns true.
func RefCompare[T any](a, b T, compare func(T, T) bool) bool {
	aIsNil, bIsNil := IsNil(a), IsNil(b)
	if aIsNil && bIsNil {
		return true
	}
	if aIsNil || bIsNil {
		return false
	}
	return compare(a, b)
}

// Less returns true if a is less than b, for any ordered type T.
func Less[T constraints.Ordered](a, b T) bool {
	return a < b
}

// IsNil returns true when the given value is nil. If it's a reference kind with
// a dynamic type, then we look for the dynamic value being nil. This is not
// required for this project at the time of this writing, due to value construction,
// but it's good for future-proofing against what could become very subtle errors.
// This is a consequence of how interfaces work in Go.
func IsNil(a interface{}) bool {
	if a == nil {
		return true
	}
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Interface, reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
