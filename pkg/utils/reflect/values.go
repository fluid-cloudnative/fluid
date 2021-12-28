package reflect

import (
	"fmt"
	"reflect"
	ref "reflect"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils/slice"
	corev1 "k8s.io/api/core/v1"
)

// valueByTypeSearcher provides valueByType for find the originalValue name by using struct type
type valueByTypeSearcher struct {
	targetNames map[string]ref.Value
}

func newValueByTypeSearcher() *valueByTypeSearcher {
	return &valueByTypeSearcher{
		targetNames: map[string]ref.Value{},
	}
}

func ValueByType(original interface{}, targetObject interface{}) map[string]ref.Value {
	targetType := ref.TypeOf(targetObject)
	originalValue := ref.ValueOf(original)
	return newValueByTypeSearcher().valueByType(originalValue, "RootObject", targetType)
}

func (f *valueByTypeSearcher) valueByType(currentValue ref.Value, currentName string, targetType ref.Type) map[string]ref.Value {

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Info("Reflection: Failed to set", "name", currentName)
	// 	}
	// }()

	currentValueStr := currentValue.String()
	log.V(1).Info("valueByType enter", "currentValue", currentValueStr, "type", currentValue.Type(), "currentName", currentName, "targetNames", f.targetNames)

	// If the target is matched
	if currentValue.Type() == targetType {
		f.targetNames[currentName] = currentValue
		return f.targetNames
	}

	switch currentValue.Kind() {
	// The first cases handle nested structures and search them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case ref.Ptr, ref.Interface:
		currentName = currentValue.Type().Elem().Name()
		if !currentValue.Elem().IsValid() {
			if currentValue.CanSet() {
				currentValue.Set(reflect.New(currentValue.Type().Elem()))
			} else {
				return f.targetNames
			}
		}

		f.valueByType(currentValue.Elem(), currentValue.Type().Elem().Name(), targetType)
	case ref.Struct:
		for i := 0; i < currentValue.NumField(); i += 1 {
			field := currentValue.Field(i)
			name := currentValue.Type().Field(i).Name
			f.valueByType(field, name, targetType)
		}
	}

	log.V(1).Info("valueByType exit", "currentValue", currentValue.String(), "currentName", currentName, "targetNames", f.targetNames)

	return f.targetNames
}

func valueFromObject(object interface{}, searchObject interface{}, nominateName string, excludeMatches []string) (name string, value ref.Value, err error) {
	names := ValueByType(object, searchObject)
	nameKeys := make([]string, len(names))

	i := 0
	for k := range names {
		nameKeys[i] = k
		i++
	}

	namesToExclude := []string{}

	// 1. Prefer nominateName if it provides
	if len(nominateName) != 0 {
		if slice.ContainsString(nameKeys, nominateName) {
			name = nominateName
			return name, names[name], nil
		}
	}

	// 2. Filter out exclude name
	for _, match := range excludeMatches {
		for _, nameToExclude := range nameKeys {
			if len(match) == 0 {
				continue
			}
			if strings.Contains(
				strings.ToLower(nameToExclude),
				strings.ToLower(match),
			) {
				namesToExclude = append(namesToExclude, nameToExclude)
			}
		}
	}

	for _, exclude := range namesToExclude {
		nameKeys = slice.RemoveString(nameKeys, exclude)
	}

	// 3. Checkout what's in names, if there are elements more than 1, skip it
	if len(names) == 1 {
		name = nameKeys[0]
		return name, names[name], nil
	}

	// 4. Checkout what's in names, if there are elements more than 1, return error
	return name, ref.Value{}, fmt.Errorf("can't determine the names in %v", names)
}

// ContainersValueFromObject gets the name of originalValue with the containers slice
func ContainersValueFromObject(object interface{}, nominateName string, excludeMatches []string) (name string, value ref.Value, err error) {
	return valueFromObject(object, []corev1.Container{}, nominateName, excludeMatches)
}

// VolumesValueFromObject gets the name of originalValue with the containers slice
func VolumesValueFromObject(object interface{}, nominateName string, excludeMatches []string) (name string, value ref.Value, err error) {
	return valueFromObject(object, []corev1.Volume{}, nominateName, excludeMatches)
}
