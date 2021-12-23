package utils

import (
	"reflect"
)

type fieldNameByTypeSearcher struct {
	targetNames []string
}

func FieldNameByType(original interface{}, targetObject interface{}) (targetNames []string) {
	targetType := reflect.TypeOf(targetObject)
	originalType := reflect.TypeOf(original)
	search := &fieldNameByTypeSearcher{}
	targetNames = search.fieldNameByType(originalType, "", targetType)
	return
}

func (f *fieldNameByTypeSearcher) fieldNameByType(currentType reflect.Type, currentName string, targetType reflect.Type) (targetNames []string) {

	log.V(1).Info("fieldNameByType enter", "currentType", currentType.String(), "currentName", currentName, "targetNames", f.targetNames)

	switch currentType.Kind() {
	// The first cases handle nested structures and search them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case reflect.Ptr, reflect.Interface:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalType := currentType.Elem()
		f.fieldNameByType(originalType, currentName, targetType)

	case reflect.Slice:
		originalType := currentType.Elem()
		f.fieldNameByType(originalType, currentName, targetType)

	// If it is a struct we serarch each field
	case reflect.Struct:
		for i := 0; i < currentType.NumField(); i += 1 {
			field := currentType.Field(i)
			if field.Type == targetType {
				f.targetNames = append(f.targetNames, field.Name)
			} else {
				f.fieldNameByType(field.Type, field.Name, targetType)
			}
		}
	default:
		if currentType == targetType {
			f.targetNames = append(f.targetNames, currentName)
		}
	}

	log.V(1).Info("fieldNameByType exit", "currentType", currentType.String(), "currentName", currentName, "targetNames", f.targetNames)

	return f.targetNames
}
