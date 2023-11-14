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

package jindo

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCheckMasterReady(t *testing.T) {
	statefulsetInputs := []v1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	JindoRuntimeInputs := []datav1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 1,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				APIGatewayStatus: &datav1alpha1.APIGatewayStatus{
					Endpoint: "test-endpoint",
				},
			},
		},
	}
	for _, JindoRuntime := range JindoRuntimeInputs {
		testObjs = append(testObjs, JindoRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JindoEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine         JindoEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: false,
		},
		{
			engine:         engines[1],
			expectedResult: false,
		},
	}

	for _, test := range testCases {
		if ready, _ := test.engine.CheckMasterReady(); ready != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
		if !test.expectedResult {
			continue
		}
		JindoRuntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get runtime %v", err)
			return
		}
		if len(JindoRuntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime conditions")
			return
		}
	}
}

func TestShouldSetupMaster(t *testing.T) {
	JindoRuntimeInputs := []datav1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNone,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, JindoRuntime := range JindoRuntimeInputs {
		testObjs = append(testObjs, JindoRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JindoEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
		},
	}

	var testCases = []struct {
		engine         JindoEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: false,
		},
		{
			engine:         engines[1],
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		if should, _ := test.engine.ShouldSetupMaster(); should != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
	}
}

func TestSetupMaster(t *testing.T) {
	statefulSetInputs := []v1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulSet := range statefulSetInputs {
		testObjs = append(testObjs, statefulSet.DeepCopy())
	}

	JindoRuntimeInputs := []datav1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, JindoRuntime := range JindoRuntimeInputs {
		testObjs = append(testObjs, JindoRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JindoEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine                JindoEngine
		expectedSelector      string
		expectedConfigMapName string
	}{
		{
			engine:                engines[0],
			expectedConfigMapName: "spark--values",
			expectedSelector:      "app=jindo,release=spark,role=jindo-worker",
		},
	}

	for _, test := range testCases {
		_ = test.engine.SetupMaster()
		JindoRuntime, _ := test.engine.getRuntime()
		if len(JindoRuntime.Status.Conditions) != 0 {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}
