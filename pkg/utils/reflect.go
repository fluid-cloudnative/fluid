package utils

import (
	"reflect"
)

type fieldNameByTypeSearcher struct {
	targetNames map[string]bool
}

func FieldNameByType(original interface{}, targetObject interface{}) (targetNames []string) {
	targetType := reflect.TypeOf(targetObject)
	originalType := reflect.TypeOf(original)
	keys := NewfieldNameByTypeSearcher().fieldNameByType(originalType, "", targetType)
	for key, _ := range keys {
		targetNames = append(targetNames, key)
	}
	return
}

func NewfieldNameByTypeSearcher() *fieldNameByTypeSearcher {
	return &fieldNameByTypeSearcher{
		targetNames: map[string]bool{},
	}
}

func (f *fieldNameByTypeSearcher) fieldNameByType(currentType reflect.Type, currentName string, targetType reflect.Type) map[string]bool {
	currentTypeStr := currentType.String()
	log.V(1).Info("fieldNameByType enter", "currentType", currentTypeStr, "currentName", currentName, "targetNames", f.targetNames)

	switch currentType.Kind() {
	// The first cases handle nested structures and search them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case reflect.Ptr, reflect.Interface, reflect.Slice:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalType := currentType.Elem()
		if originalType == targetType {
			// f.targetNames = append(f.targetNames, currentName)
			f.targetNames[currentName] = true
		} else {
			f.fieldNameByType(originalType, currentName, targetType)
		}
	// If it is a struct we serarch each field
	case reflect.Struct:
		for i := 0; i < currentType.NumField(); i += 1 {
			field := currentType.Field(i)
			if field.Type == targetType {
				f.targetNames[field.Name] = true
			} else {
				f.fieldNameByType(field.Type, field.Name, targetType)
			}
		}
	default:
		if currentType == targetType {
			f.targetNames[currentName] = true
		}
	}

	log.V(1).Info("fieldNameByType exit", "currentType", currentType.String(), "currentName", currentName, "targetNames", f.targetNames)

	return f.targetNames
}
