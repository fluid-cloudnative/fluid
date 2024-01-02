/*
Copyright 2024 The Fluid Authors.
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

package vineyard

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

	vineyardruntimeInputs := []datav1alpha1.VineyardRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Master: datav1alpha1.MasterSpec{
					VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
						Replicas: 1,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Master: datav1alpha1.MasterSpec{
					VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
						Replicas: 2,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Master: datav1alpha1.MasterSpec{
					VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
						Replicas: 1,
					},
				},
			},
		},
	}
	for _, vineyardruntime := range vineyardruntimeInputs {
		testObjs = append(testObjs, vineyardruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []VineyardEngine{
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
		engine         VineyardEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: true,
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
		vineyardruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get runtime %v", err)
			return
		}
		if len(vineyardruntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime conditions")
			return
		}
	}
}

func TestShouldSetupMaster(t *testing.T) {
	vineyardruntimeInputs := []datav1alpha1.VineyardRuntime{
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
	for _, vineyardruntime := range vineyardruntimeInputs {
		testObjs = append(testObjs, vineyardruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []VineyardEngine{
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
		engine         VineyardEngine
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

	vineyardruntimeInputs := []datav1alpha1.VineyardRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, vineyardruntime := range vineyardruntimeInputs {
		testObjs = append(testObjs, vineyardruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []VineyardEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine                VineyardEngine
		expectedSelector      string
		expectedConfigMapName string
	}{
		{
			engine:                engines[0],
			expectedConfigMapName: "spark--values",
			expectedSelector:      "app=vineyard,release=spark,role=vineyard-worker",
		},
	}

	for _, test := range testCases {
		if err := test.engine.SetupMaster(); err != nil {
			t.Errorf("fail to exec the func with error %v", err)
			return
		}
		vineyardruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime")
			return
		}
		if vineyardruntime.Status.Selector != test.expectedSelector || vineyardruntime.Status.ValueFileConfigmap != test.expectedConfigMapName ||
			len(vineyardruntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}
