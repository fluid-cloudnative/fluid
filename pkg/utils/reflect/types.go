package reflect

import (
	"fmt"
	ref "reflect"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils/slice"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("reflect")
}

// fieldNameByTypeSearcher provides fieldNameByType for find the field name by using struct type
type fieldNameByTypeSearcher struct {
	targetNames map[string]bool
}

func FieldNameByType(original interface{}, targetObject interface{}) (targetNames []string) {
	targetType := ref.TypeOf(targetObject)
	originalType := ref.TypeOf(original)
	keys := NewfieldNameByTypeSearcher().fieldNameByType(originalType, "", targetType)
	for key := range keys {
		targetNames = append(targetNames, key)
	}
	return
}

func NewfieldNameByTypeSearcher() *fieldNameByTypeSearcher {
	return &fieldNameByTypeSearcher{
		targetNames: map[string]bool{},
	}
}

func (f *fieldNameByTypeSearcher) fieldNameByType(currentType ref.Type, currentName string, targetType ref.Type) map[string]bool {
	currentTypeStr := currentType.String()
	log.V(1).Info("fieldNameByType enter", "currentType", currentTypeStr, "currentName", currentName, "targetNames", f.targetNames)

	switch currentType.Kind() {
	// The first cases handle nested structures and search them recursively

	// If it is a pointer, interface we need to unwrap and call once again
	case ref.Ptr, ref.Interface, ref.Slice:
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
	case ref.Struct:
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

func fieldNameFromObject(object interface{}, searchObject interface{}, nominateName string, excludeMatches []string) (name string, err error) {
	names := FieldNameByType(object, searchObject)
	namesToExclude := []string{}

	// 1. Prefer nominateName if it provides
	if len(nominateName) != 0 {
		if slice.ContainsString(names, nominateName) {
			name = nominateName
			return
		}
	}

	// 2. Filter out exclude name
	for _, match := range excludeMatches {
		for _, nameToExclude := range names {
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
		names = slice.RemoveString(names, exclude)
	}

	// 3. Checkout what's in names, if there are elements more than 1, skip it
	if len(names) == 1 {
		name = names[0]
		return
	}

	// 4. Checkout what's in names, if there are elements more than 1, return error
	return name, fmt.Errorf("can't determine the names in %v", names)
}

// ContainersFieldNameFromObject gets the name of field with the containers slice
func ContainersFieldNameFromObject(object interface{}, nominateName string, excludeMatches []string) (name string, err error) {
	return fieldNameFromObject(object, []corev1.Container{}, nominateName, excludeMatches)
}

// VolumesFieldNameFromObject gets the name of field with the containers slice
func VolumesFieldNameFromObject(object interface{}, nominateName string, excludeMatches []string) (name string, err error) {
	return fieldNameFromObject(object, []corev1.Volume{}, nominateName, excludeMatches)
}
