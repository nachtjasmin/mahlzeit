package testhelper

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// PartialEqual takes a partially-initialized struct its expected (first
// argument) side, and compares all non-zero values on it or any substructs to
// the symmetrical values of actual (second argument). In other words, it can
// take a struct that only has values that the caller cares about, and compares
// just those, even if actual has other non-zero values.
//
// Example use:
//
//	testhelper.PartialEqual(t, dbsqlc.Cluster{
//	    Environment: sql.NullString{String: string(dbsqlc.ClusterEnvironmentProduction), Valid: true},
//	    Name:        cluster.Name,
//	}, cluster)
//
// WARNING: In Go, there's no difference between an explicit zero value versus an
// implicit one from when a field is just left out of a struct initialization,
// so there's no way for this helper to compare zero values set on the expected
// side -- they'll just be silently ignored. So watch out for use of anything
// like `false`, `nil`, `sql.NullString{}`, etc. as none of them will work.
// Recommended work around is to compare non-zero values with this assertion,
// and then assert on values expected to be zero with `require.Zero`. For API
// resources, also consider making fields pointer primitives like `*bool` or
// `*int` so that an explicit `ptrutil.Ptr(false)` or `ptrutil.Ptr(0)` can be
// checked.
//
// This is taken and adapted from: https://brandur.org/fragments/partial-equal
func PartialEqual[T any](t testing.TB, expected, actual T) {
	t.Helper()

	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	var wasPtr bool
	if expectedVal.Kind() == reflect.Ptr {
		if actualVal.Kind() != reflect.Ptr {
			panic("expected value is a pointer; actual value should also be a pointer")
		}

		expectedVal = expectedVal.Elem()
		actualVal = actualVal.Elem()

		wasPtr = true
	}

	switch {
	case expectedVal.Kind() != reflect.Struct:
		panic("expected value must be a struct")

	case !wasPtr && actualVal.Kind() == reflect.Ptr:
		panic("expected value was not a pointer; actual value should also not be a pointer")

	case expectedVal.Type() != actualVal.Type():
		panic("expected and actual values must be the same type")
	}

	partialActual := buildPartialStruct(expectedVal, actualVal)

	if !wasPtr {
		partialActual = reflect.ValueOf(partialActual).Elem().Interface()
	}

	assert.Equal[any](t, expected, partialActual, "Expected all non-zero fields on structs to be the same")
}

// Builds a partial struct, taking values from actualVal according to which
// values are non-zero in expectedVal.
func buildPartialStruct(expectedVal, actualVal reflect.Value) any {
	// Creates a pointer to a new value of the given type.
	partialActual := reflect.New(actualVal.Type())

	for i := 0; i < expectedVal.NumField(); i++ {
		expectedField := expectedVal.Field(i)
		actualField := actualVal.Field(i)

		if expectedField.IsZero() {
			continue
		}

		field := partialActual.Elem().Field(i)

		switch {
		case expectedField.Kind() == reflect.Struct:
			if !actualField.IsZero() {
				s := buildPartialStruct(expectedField, actualField)
				field.Set(reflect.ValueOf(s).Elem())
			}

		case expectedField.Kind() == reflect.Ptr && expectedField.Elem().Kind() == reflect.Struct:
			if !actualField.IsNil() {
				s := buildPartialStruct(expectedField.Elem(), actualField.Elem())
				field.Set(reflect.ValueOf(s))
			}

		default:
			field.Set(actualField)
		}
	}

	return partialActual.Interface()
}
