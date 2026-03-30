/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

type TestCase struct {
	name            string
	namespace       string
	withRuntimeInfo bool
	isDeleted       bool
	isErr           bool
}

func newTestEFCEngine(client client.Client, name string, namespace string, withRuntimeInfo bool) *EFCEngine {
	runTime := &datav1alpha1.EFCRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.EFCRuntime)
	if !withRuntimeInfo {
		runTimeInfo = nil
		runTime = nil
	}
	engine := &EFCEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}

	return engine
}

func doTestCases(testCases []TestCase, resources []runtime.Object, t *testing.T) {
	for _, test := range testCases {
		fakeClient := fake.NewFakeClientWithScheme(testScheme, resources...)
		engine := newTestEFCEngine(fakeClient, test.name, test.namespace, test.withRuntimeInfo)
		err := engine.DeleteVolume(context.Background())

		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("%s/%s withRuntimeInfo=%t: expected error=%t, got %t.", test.namespace, test.name, test.withRuntimeInfo, test.isErr, isErr)
		}

		pv := &v1.PersistentVolume{}
		nullPV := v1.PersistentVolume{}
		keyPV := types.NamespacedName{
			Name: fmt.Sprintf("%s-%s", engine.namespace, engine.name),
		}
		_ = engine.Client.Get(context.TODO(), keyPV, pv)
		if test.isDeleted != reflect.DeepEqual(nullPV, *pv) {
			t.Errorf("%s/%s withRuntimeInfo=%t: PV still exist after delete.", test.namespace, test.name, test.withRuntimeInfo)
		}

		pvc := &v1.PersistentVolumeClaim{}
		nullPVC := v1.PersistentVolumeClaim{}
		keyPVC := types.NamespacedName{
			Name:      engine.name,
			Namespace: engine.namespace,
		}
		_ = engine.Client.Get(context.TODO(), keyPVC, pvc)
		if test.isDeleted != reflect.DeepEqual(nullPVC, *pvc) {
			t.Errorf("%s/%s withRuntimeInfo=%t: PVC still exist after delete.", test.namespace, test.name, test.withRuntimeInfo)
		}
	}
}

func TestEFCEngine_DeleteVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-efcdemo",
				//Namespace:   "fluid",
				Annotations: common.GetExpectedFluidAnnotations(),
			},
			Spec: v1.PersistentVolumeSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-error",
				//Namespace:   "fluid",
				Annotations: common.GetExpectedFluidAnnotations(),
			},
			Spec: v1.PersistentVolumeSpec{},
		},
	}

	tests := []runtime.Object{}

	for _, pvInput := range testPVInputs {
		tests = append(tests, pvInput.DeepCopy())
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "error",
				Namespace:   "fluid",
				Annotations: common.GetExpectedFluidAnnotations(),
				Finalizers:  []string{"kubernetes.io/pvc-protection"}, // err because it needs sleep
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},
	}

	for _, pvcInput := range testPVCInputs {
		tests = append(tests, pvcInput.DeepCopy())
	}

	var testCases = []TestCase{
		{
			name:            "efcdemo",
			namespace:       "fluid",
			withRuntimeInfo: true,
			isDeleted:       true,
			isErr:           false,
		},
		{
			name:            "error",
			namespace:       "fluid",
			withRuntimeInfo: true,
			isDeleted:       false,
			isErr:           true,
		},
		{
			name:            "efcdemo",
			namespace:       "fluid",
			withRuntimeInfo: false,
			isDeleted:       true,
			isErr:           true,
		},
	}
	doTestCases(testCases, tests, t)
}
