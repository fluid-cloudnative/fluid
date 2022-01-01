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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func TestLocateContainers(t *testing.T) {
	type testCase struct {
		name    string
		content string
		expect  []common.Anchor
	}

	testcases := []testCase{
		{
			name:    "statefulset",
			content: stsYaml,
			expect: []common.Anchor{
				UnstructuredAnchor{}, UnstructuredAnchor{},
			},
		},
		{
			name:    "tfjob",
			content: tfjobYaml,
			expect: []common.Anchor{
				UnstructuredAnchor{}, UnstructuredAnchor{},
			},
		}, {
			name:    "pytorch",
			content: pytorchYaml,
			expect: []common.Anchor{
				UnstructuredAnchor{}, UnstructuredAnchor{},
			},
		}, {
			name:    "argo",
			content: argoYaml,
			expect: []common.Anchor{
				UnstructuredAnchor{}, UnstructuredAnchor{},
			},
		}, {
			name:    "spark",
			content: sparkYaml,
			expect: []common.Anchor{
				UnstructuredAnchor{}, UnstructuredAnchor{},
			},
		},
	}

	for _, testcase := range testcases {
		// got := ValueByType(testcase.original, testcase.target)

		// result := differenceMap(got, testcase.expect)
		// if len(result) > 0 {
		// 	t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		// }
		obj := &unstructured.Unstructured{}

		dec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, gvk, err := dec.Decode([]byte(testcase.content), nil, obj)
		if err != nil {
			t.Errorf("Failed to decode due to %v and gvk is %v", err, gvk)
		}

		app := NewUnstructuredApplication(obj)
		got, err := app.LocateContainers()
		if err != nil {
			t.Errorf("testcase %s failed due to error %v", testcase.name, err)
		}

		if !reflect.DeepEqual(got, testcase.expect) {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}

	}

}

func TestInjectObjectForUnstructed(t *testing.T) {

	obj := &unstructured.Unstructured{}

	dec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(tfjobYaml), nil, obj)
	if err != nil {
		t.Errorf("Failed to decode due to %v", err)
	}

	// Get the common metadata, and show GVK
	fmt.Println(obj.GetName(), gvk.String())

	app := NewUnstructuredApplication(obj)
	ans, err := app.LocateVolumes()
	if err != nil {
		t.Errorf("Failed to LocateVolumes due to %v", err)
	}
	fmt.Printf("ans:%v", ans)
	ans, err = app.LocateContainers()
	if err != nil {
		t.Errorf("Failed to LocateVolumes due to %v", err)
	}
	fmt.Printf("ans:%v", ans)
	out := app.GetObject()
	if err != nil {
		t.Errorf("Failed to GetObject due to %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(out)

}
