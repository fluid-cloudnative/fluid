/*
Copyright 2021 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unstructured

import (
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// UnstructuredPointer is used to locate the path
// inside the unstructured
type UnstructuredPointer struct {
	fields []string
}

func NewUnstructuredPointerFromPath(path string) common.Pointer {
	fields := strings.Split(path, "/")
	return NewUnstructuredPointer(fields, "")
}

func NewUnstructuredPointer(fields []string, end string) common.Pointer {
	fieldsToAdd := []string{}
	if len(end) > 0 {
		for _, field := range fields {
			fieldsToAdd = append(fieldsToAdd, field)
			if field == end {
				break
			}
		}
	} else {
		fieldsToAdd = fields
	}

	return UnstructuredPointer{
		fields: fieldsToAdd,
	}
}

func (a UnstructuredPointer) Key() (id string) {
	return strings.Join(a.fields, "/")
}

func (a UnstructuredPointer) Paths() (paths []string) {
	return a.fields
}

func (a UnstructuredPointer) String() string {
	return a.Key()
}

func (a UnstructuredPointer) Parent() (p common.Pointer, err error) {
	if len(a.fields) > 0 {
		fields := a.fields
		p = NewUnstructuredPointer(fields[:len(fields)-1], "")
	} else {
		return nil, fmt.Errorf("failed to find parent from %v", a.fields)
	}

	return
}

func (a UnstructuredPointer) Child(name string) (p common.Pointer) {
	// fields := []string{}
	fields := append([]string{}, a.fields...)
	return NewUnstructuredPointer(append(fields, name), "")
}
