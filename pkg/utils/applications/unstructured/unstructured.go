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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/nqd/flat"
)

const (
	delimiter            string = ":"
	containersMatchStr   string = "containers:0:volumeMounts:0"
	containersEndStr     string = "containers"
	volumesMatchStr      string = "volumes:0"
	volumesEndStr        string = "volumes"
	volumeMountsMatchStr string = "volumeMounts:0"
	volumeMountssEndStr  string = "volumeMounts"
)

var (
	defaultContainersName string = "containers"
	defaultVolumessName   string = "volumes"
)

// UnstructuredApp allows objects that do not have Golang structs registered to be manipulated
// generically. This can be used to deal with the API objects from a plug-in. UnstructuredApp
// objects can handle the common object like Container, Volume
type UnstructuredApplication struct {
	root *unstructured.Unstructured
}

type UnstructuredApplicationPodSpec struct {
	*unstructuredObject
	root *unstructured.Unstructured
	// absolute Ptr
	ptr common.Pointer
	// relative Ptr
	containersPtr common.Pointer
	volumesPtr    common.Pointer
}

func NewUnstructuredApplicationPodSpec(root *unstructured.Unstructured, ptr common.Pointer, containersName, volumesName *string) (spec *UnstructuredApplicationPodSpec, err error) {
	field, found, err := unstructured.NestedFieldCopy(root.Object, ptr.Paths()...)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("failed to find the volumes from %v", ptr.Paths())
	}

	original, ok := field.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse %v", field)
	}
	newRoot := unstructured.Unstructured{Object: original}

	if containersName == nil {
		containersName = &defaultContainersName
	}

	if volumesName == nil {
		volumesName = &defaultVolumessName
	}

	spec = &UnstructuredApplicationPodSpec{
		root:               &newRoot,
		ptr:                ptr,
		containersPtr:      ptr.Child(*containersName),
		volumesPtr:         ptr.Child(*volumesName),
		unstructuredObject: &unstructuredObject{},
	}

	return
}

func NewUnstructuredApplication(obj *unstructured.Unstructured) common.Application {
	return &UnstructuredApplication{
		root: obj,
	}
}

func (u *UnstructuredApplication) GetPodSpecs() (specs []common.Object, err error) {
	volumePtrs, err := u.LocateVolumes()
	if err != nil {
		return
	}

	specs = make([]common.Object, 0, len(volumePtrs))
	for _, volumePtr := range volumePtrs {
		ptr, err := volumePtr.Parent()
		if err != nil {
			return nil, err
		}
		spec, err := NewUnstructuredApplicationPodSpec(
			u.root,
			ptr,
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return
}

func (u *UnstructuredApplication) SetPodSpecs(specs []common.Object) (err error) {
	return
}

func (u *UnstructuredApplication) GetObject() (obj runtime.Object) {
	return u.root
}

func (u *UnstructuredApplication) LocateContainers() (pointers []common.Pointer, err error) {
	return u.locate(containersMatchStr, containersEndStr)
}

func (u *UnstructuredApplication) LocateVolumes() (pointers []common.Pointer, err error) {
	return u.locate(volumesMatchStr, volumesEndStr)
}

func (u *UnstructuredApplication) LocateVolumeMounts() (pointers []common.Pointer, err error) {
	return u.locate(volumeMountsMatchStr, volumeMountssEndStr)
}

func (u *UnstructuredApplication) locate(matchStr, endStr string) (pointers []common.Pointer, err error) {
	pointersMap := map[string]bool{}
	out, err := flat.Flatten(u.root.Object, &flat.Options{
		Delimiter: delimiter,
	})
	if err != nil {
		return pointers, err
	}
	for key := range out {
		if strings.Contains(key, matchStr) {
			anchor := NewUnstructuredPointer(strings.Split(key, ":"), endStr)
			if _, found := pointersMap[anchor.Key()]; !found {
				pointers = append(pointers, anchor)
				pointersMap[anchor.Key()] = true
			}
		}
	}
	return pointers, err
}
