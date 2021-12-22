package utils

import (
	"fmt"
	"reflect"
)

func FieldNameByType(original interface{}, targetObject interface{}) (targetNames []string, found bool) {
	targetNames = []string{}
	targetType := reflect.TypeOf(targetObject)
	originalType := reflect.TypeOf(original)
	found = fieldNameByType(originalType, targetNames, targetType)
	return
}


func fieldNameByType(original reflect.Type, targetNames []string, targetType reflect.Type) (match bool) {

	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case reflect.Ptr, reflect.Slice, reflect.Interface:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalType := original.Elem()
		return fieldNameByType(originalType, targetNames, targetType)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			field := original.Field(i)
			if field.Type == targetType {
				targetNames = append(targetNames, field.Name)
			} else {
				return fieldNameByType(field.Type, targetNames, targetType)
			}
		}
	}

	return match
}
