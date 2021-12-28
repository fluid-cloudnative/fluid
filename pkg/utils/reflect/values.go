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
	currentValueStr := currentValue.String()
	log.V(1).Info("valueByType enter", "currentValue", currentValueStr, "type", currentValue.Type(), "currentName", currentName, "targetNames", f.targetNames)

	// If the target is matched
	if currentValue.Type() == targetType {
		f.targetNames[currentName] = currentValue
	}

	switch currentValue.Kind() {
	// The first cases handle nested structures and search them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case ref.Ptr, ref.Interface:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := currentValue.Elem()

		if !originalValue.IsValid() {
			// fmt.Printf("result isZero %v\n", originalValue.IsZero())
			// fmt.Printf("result Kind %v\n", originalValue.Kind())
			// fmt.Printf("result isNil %v\n", originalValue.IsNil())
			// fmt.Printf("result %v\n", originalValue.Type().String())
			// fmt.Printf("result %v\n", originalValue.CanSet())
			// fmt.Printf("result %v\n", originalValue.CanAddr())
			originalValue.Set(reflect.New(currentValue.Type().Elem()))
		}

		if originalValue.Type() != targetType {
			// f.targetNames = append(f.targetNames, currentName)
			f.valueByType(originalValue, currentName, targetType)
		}
	case ref.Slice:
		originalValue := currentValue.Elem()
		if !originalValue.IsValid() {
			elemType := reflect.TypeOf(currentValue.Type().Elem())
			slice := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 10)
			originalValue.Set(slice)
		}

		if originalValue.Type() != targetType {
			f.valueByType(originalValue, originalValue.Type().Name(), targetType)
		}

	// If it is a struct we serarch each originalValue
	case ref.Struct:
		for i := 0; i < currentValue.NumField(); i += 1 {
			originalValue := currentValue.Field(i)
			if originalValue.Type() != targetType {
				f.valueByType(originalValue, originalValue.Type().Name(), targetType)
			}
		}
	}

	if currentValue.Type() == targetType {
		f.targetNames[currentName] = currentValue
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
