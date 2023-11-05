package assertions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/smarty/assertions/internal/go-render/render"
)

// ShouldEqual receives exactly two parameters and does an equality check
// using the following semantics:
// It uses reflect.DeepEqual in most cases, but also compares numerics
// regardless of specific type and compares time.Time values using the
// time.Equal method.
func ShouldEqual(actual any, expected ...any) string {
	if message := need(1, expected); message != success {
		return message
	}
	return shouldEqual(actual, expected[0])
}
func shouldEqual(actual, expected any) string {
	for _, spec := range equalitySpecs {
		if !spec.assertable(actual, expected) {
			continue
		}
		if spec.passes(actual, expected) {
			return success
		}
		break
	}
	renderedExpected, renderedActual := render.Render(expected), render.Render(actual)
	if renderedActual == renderedExpected {
		message := fmt.Sprintf(shouldHaveResembledButTypeDiff, renderedExpected, renderedActual)
		return serializer.serializeDetailed(expected, actual, message)
	}
	message := fmt.Sprintf(shouldHaveResembled, renderedExpected, renderedActual) +
		composePrettyDiff(renderedExpected, renderedActual)
	return serializer.serializeDetailed(expected, actual, message)
}

// ShouldNotEqual receives exactly two parameters and does an inequality check.
// See ShouldEqual for details on how equality is determined.
func ShouldNotEqual(actual any, expected ...any) string {
	if fail := need(1, expected); fail != success {
		return fail
	} else if ShouldEqual(actual, expected[0]) == success {
		return fmt.Sprintf(shouldNotHaveBeenEqual, actual, expected[0])
	}
	return success
}

// ShouldAlmostEqual makes sure that two parameters are close enough to being equal.
// The acceptable delta may be specified with a third argument,
// or a very small default delta will be used.
func ShouldAlmostEqual(actual any, expected ...any) string {
	actualFloat, expectedFloat, deltaFloat, err := cleanAlmostEqualInput(actual, expected...)

	if err != "" {
		return err
	}

	if math.Abs(actualFloat-expectedFloat) <= deltaFloat {
		return success
	} else {
		return fmt.Sprintf(shouldHaveBeenAlmostEqual, actualFloat, expectedFloat)
	}
}

// ShouldNotAlmostEqual is the inverse of ShouldAlmostEqual
func ShouldNotAlmostEqual(actual any, expected ...any) string {
	actualFloat, expectedFloat, deltaFloat, err := cleanAlmostEqualInput(actual, expected...)

	if err != "" {
		return err
	}

	if math.Abs(actualFloat-expectedFloat) > deltaFloat {
		return success
	} else {
		return fmt.Sprintf(shouldHaveNotBeenAlmostEqual, actualFloat, expectedFloat)
	}
}

func cleanAlmostEqualInput(actual any, expected ...any) (float64, float64, float64, string) {
	deltaFloat := 0.0000000001

	if len(expected) == 0 {
		return 0.0, 0.0, 0.0, "This assertion requires exactly one comparison value and an optional delta (you provided neither)"
	} else if len(expected) == 2 {
		delta, err := getFloat(expected[1])

		if err != nil {
			return 0.0, 0.0, 0.0, "The delta value " + err.Error()
		}

		deltaFloat = delta
	} else if len(expected) > 2 {
		return 0.0, 0.0, 0.0, "This assertion requires exactly one comparison value and an optional delta (you provided more values)"
	}

	actualFloat, err := getFloat(actual)
	if err != nil {
		return 0.0, 0.0, 0.0, "The actual value " + err.Error()
	}

	expectedFloat, err := getFloat(expected[0])
	if err != nil {
		return 0.0, 0.0, 0.0, "The comparison value " + err.Error()
	}

	return actualFloat, expectedFloat, deltaFloat, ""
}

// returns the float value of any real number, or error if it is not a numerical type
func getFloat(num any) (float64, error) {
	numValue := reflect.ValueOf(num)
	numKind := numValue.Kind()

	if numKind == reflect.Int ||
		numKind == reflect.Int8 ||
		numKind == reflect.Int16 ||
		numKind == reflect.Int32 ||
		numKind == reflect.Int64 {
		return float64(numValue.Int()), nil
	} else if numKind == reflect.Uint ||
		numKind == reflect.Uint8 ||
		numKind == reflect.Uint16 ||
		numKind == reflect.Uint32 ||
		numKind == reflect.Uint64 {
		return float64(numValue.Uint()), nil
	} else if numKind == reflect.Float32 ||
		numKind == reflect.Float64 {
		return numValue.Float(), nil
	} else {
		return 0.0, errors.New("must be a numerical type, but was: " + numKind.String())
	}
}

// ShouldEqualJSON receives exactly two parameters and does an equality check by marshalling to JSON
func ShouldEqualJSON(actual any, expected ...any) string {
	if message := need(1, expected); message != success {
		return message
	}

	expectedString, expectedErr := remarshal(expected[0].(string))
	if expectedErr != nil {
		return "Expected value not valid JSON: " + expectedErr.Error()
	}

	actualString, actualErr := remarshal(actual.(string))
	if actualErr != nil {
		return "Actual value not valid JSON: " + actualErr.Error()
	}

	return ShouldEqual(actualString, expectedString)
}
func remarshal(value string) (string, error) {
	var structured any
	err := json.Unmarshal([]byte(value), &structured)
	if err != nil {
		return "", err
	}
	canonical, _ := json.Marshal(structured)
	return string(canonical), nil
}

// ShouldResemble is an alias for ShouldEqual.
func ShouldResemble(actual any, expected ...any) string {
	return ShouldEqual(actual, expected...)
}

// ShouldNotResemble is an alias for ShouldNotEqual.
func ShouldNotResemble(actual any, expected ...any) string {
	return ShouldNotEqual(actual, expected...)
}

// ShouldPointTo receives exactly two parameters and checks to see that they point to the same address.
func ShouldPointTo(actual any, expected ...any) string {
	if message := need(1, expected); message != success {
		return message
	}
	return shouldPointTo(actual, expected[0])

}
func shouldPointTo(actual, expected any) string {
	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)

	if ShouldNotBeNil(actual) != success {
		return fmt.Sprintf(shouldHaveBeenNonNilPointer, "first", "nil")
	} else if ShouldNotBeNil(expected) != success {
		return fmt.Sprintf(shouldHaveBeenNonNilPointer, "second", "nil")
	} else if actualValue.Kind() != reflect.Ptr {
		return fmt.Sprintf(shouldHaveBeenNonNilPointer, "first", "not")
	} else if expectedValue.Kind() != reflect.Ptr {
		return fmt.Sprintf(shouldHaveBeenNonNilPointer, "second", "not")
	} else if ShouldEqual(actualValue.Pointer(), expectedValue.Pointer()) != success {
		actualAddress := reflect.ValueOf(actual).Pointer()
		expectedAddress := reflect.ValueOf(expected).Pointer()
		return serializer.serialize(expectedAddress, actualAddress, fmt.Sprintf(shouldHavePointedTo,
			actual, actualAddress,
			expected, expectedAddress))
	}
	return success
}

// ShouldNotPointTo receives exactly two parameters and checks to see that they point to different addresess.
func ShouldNotPointTo(actual any, expected ...any) string {
	if message := need(1, expected); message != success {
		return message
	}
	compare := ShouldPointTo(actual, expected[0])
	if strings.HasPrefix(compare, shouldBePointers) {
		return compare
	} else if compare == success {
		return fmt.Sprintf(shouldNotHavePointedTo, actual, expected[0], reflect.ValueOf(actual).Pointer())
	}
	return success
}

// ShouldBeNil receives a single parameter and ensures that it is nil.
func ShouldBeNil(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	} else if actual == nil {
		return success
	} else if interfaceHasNilValue(actual) {
		return success
	}
	return fmt.Sprintf(shouldHaveBeenNil, actual)
}
func interfaceHasNilValue(actual any) bool {
	value := reflect.ValueOf(actual)
	kind := value.Kind()
	nilable := kind == reflect.Slice ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Ptr ||
		kind == reflect.Map

	// Careful: reflect.Value.IsNil() will panic unless it's an interface, chan, map, func, slice, or ptr
	// Reference: http://golang.org/pkg/reflect/#Value.IsNil
	return nilable && value.IsNil()
}

// ShouldNotBeNil receives a single parameter and ensures that it is not nil.
func ShouldNotBeNil(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	} else if ShouldBeNil(actual) == success {
		return fmt.Sprintf(shouldNotHaveBeenNil, actual)
	}
	return success
}

// ShouldBeTrue receives a single parameter and ensures that it is true.
func ShouldBeTrue(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	} else if actual != true {
		return fmt.Sprintf(shouldHaveBeenTrue, actual)
	}
	return success
}

// ShouldBeFalse receives a single parameter and ensures that it is false.
func ShouldBeFalse(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	} else if actual != false {
		return fmt.Sprintf(shouldHaveBeenFalse, actual)
	}
	return success
}

// ShouldBeZeroValue receives a single parameter and ensures that it is
// the Go equivalent of the default value, or "zero" value.
func ShouldBeZeroValue(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	}
	zeroVal := reflect.Zero(reflect.TypeOf(actual)).Interface()
	if !reflect.DeepEqual(zeroVal, actual) {
		return serializer.serialize(zeroVal, actual, fmt.Sprintf(shouldHaveBeenZeroValue, actual))
	}
	return success
}

// ShouldNotBeZeroValue receives a single parameter and ensures that it is NOT
// the Go equivalent of the default value, or "zero" value.
func ShouldNotBeZeroValue(actual any, expected ...any) string {
	if fail := need(0, expected); fail != success {
		return fail
	}
	zeroVal := reflect.Zero(reflect.TypeOf(actual)).Interface()
	if reflect.DeepEqual(zeroVal, actual) {
		return serializer.serialize(zeroVal, actual, fmt.Sprintf(shouldNotHaveBeenZeroValue, actual))
	}
	return success
}
