package assertions

import "reflect"

type equalityMethodSpecification struct{}

func (this equalityMethodSpecification) assertable(a, b any) bool {
	if !bothAreSameType(a, b) {
		return false
	}
	if !typeHasEqualMethod(a) {
		return false
	}
	if !equalMethodReceivesSameTypeForComparison(a) {
		return false
	}
	if !equalMethodReturnsBool(a) {
		return false
	}
	return true
}
func bothAreSameType(a, b any) bool {
	aType := reflect.TypeOf(a)
	if aType == nil {
		return false
	}
	if aType.Kind() == reflect.Ptr {
		aType = aType.Elem()
	}
	bType := reflect.TypeOf(b)
	return aType == bType
}
func typeHasEqualMethod(a any) bool {
	aInstance := reflect.ValueOf(a)
	equalMethod := aInstance.MethodByName("Equal")
	return equalMethod != reflect.Value{}
}
func equalMethodReceivesSameTypeForComparison(a any) bool {
	aType := reflect.TypeOf(a)
	if aType.Kind() == reflect.Ptr {
		aType = aType.Elem()
	}
	aInstance := reflect.ValueOf(a)
	equalMethod := aInstance.MethodByName("Equal")
	signature := equalMethod.Type()
	return signature.NumIn() == 1 && signature.In(0) == aType
}
func equalMethodReturnsBool(a any) bool {
	aInstance := reflect.ValueOf(a)
	equalMethod := aInstance.MethodByName("Equal")
	signature := equalMethod.Type()
	return signature.NumOut() == 1 && signature.Out(0) == reflect.TypeOf(true)
}

func (this equalityMethodSpecification) passes(A, B any) bool {
	a := reflect.ValueOf(A)
	b := reflect.ValueOf(B)
	return areEqual(a, b) && areEqual(b, a)
}
func areEqual(receiver reflect.Value, argument reflect.Value) bool {
	equalMethod := receiver.MethodByName("Equal")
	argumentList := []reflect.Value{argument}
	result := equalMethod.Call(argumentList)
	return result[0].Bool()
}
