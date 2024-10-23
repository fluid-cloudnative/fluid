package assertions

import (
	"math"
	"reflect"
	"time"
)

type specification interface {
	assertable(a, b any) bool
	passes(a, b any) bool
}

var equalitySpecs = []specification{
	numericEquality{},
	timeEquality{},
	deepEquality{},
	equalityMethodSpecification{}, // Now that we have the timeEquality spec, this could probably be removed..
}

// deepEquality compares any two values using reflect.DeepEqual.
// https://golang.org/pkg/reflect/#DeepEqual
type deepEquality struct{}

func (deepEquality) assertable(a, b any) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}
func (deepEquality) passes(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// numericEquality compares numeric values using the built-in equality
// operator (`==`). Values of differing numeric reflect.Kind are each
// converted to the type of the other and are compared with `==` in both
// directions, with one exception: two mixed integers (one signed and one
// unsigned) are always unequal in the case that the unsigned value is
// greater than math.MaxInt64. https://golang.org/pkg/reflect/#Kind
type numericEquality struct{}

func (numericEquality) assertable(a, b any) bool {
	return isNumeric(a) && isNumeric(b)
}
func (numericEquality) passes(a, b any) bool {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	if isUnsignedInteger(a) && isSignedInteger(b) && aValue.Uint() >= math.MaxInt64 {
		return false
	}
	if isSignedInteger(a) && isUnsignedInteger(b) && bValue.Uint() >= math.MaxInt64 {
		return false
	}
	aAsB := aValue.Convert(bValue.Type()).Interface()
	bAsA := bValue.Convert(aValue.Type()).Interface()
	return a == bAsA && b == aAsB
}
func isNumeric(v any) bool {
	of := reflect.TypeOf(v)
	if of == nil {
		return false
	}
	_, found := numericKinds[of.Kind()]
	return found
}
func isSignedInteger(v any) bool {
	_, found := signedIntegerKinds[reflect.TypeOf(v).Kind()]
	return found
}

var unsignedIntegerKinds = map[reflect.Kind]struct{}{
	reflect.Uint:    {},
	reflect.Uint8:   {},
	reflect.Uint16:  {},
	reflect.Uint32:  {},
	reflect.Uint64:  {},
	reflect.Uintptr: {},
}

func isUnsignedInteger(v any) bool {
	_, found := unsignedIntegerKinds[reflect.TypeOf(v).Kind()]
	return found
}

var signedIntegerKinds = map[reflect.Kind]struct{}{
	reflect.Int:   {},
	reflect.Int8:  {},
	reflect.Int16: {},
	reflect.Int32: {},
	reflect.Int64: {},
}

var numericKinds = map[reflect.Kind]struct{}{
	reflect.Int:     {},
	reflect.Int8:    {},
	reflect.Int16:   {},
	reflect.Int32:   {},
	reflect.Int64:   {},
	reflect.Uint:    {},
	reflect.Uint8:   {},
	reflect.Uint16:  {},
	reflect.Uint32:  {},
	reflect.Uint64:  {},
	reflect.Float32: {},
	reflect.Float64: {},
}

// timeEquality compares values both of type time.Time using their Equal method.
// https://golang.org/pkg/time/#Time.Equal
type timeEquality struct{}

func (timeEquality) assertable(a, b any) bool {
	return isTime(a) && isTime(b)
}
func (timeEquality) passes(a, b any) bool {
	return a.(time.Time).Equal(b.(time.Time))
}
func isTime(v any) bool {
	_, ok := v.(time.Time)
	return ok
}
