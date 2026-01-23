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

package alluxio

import (
	"os"
	"strings"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestAlluxioEngineValidate(t *testing.T) {
	os.Setenv(testutil.FluidUnitTestEnv, "true")
	defer os.Unsetenv(testutil.FluidUnitTestEnv)

	dataset, alluxioruntime := mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
	ctx := cruntime.ReconcileRequestContext{}

	testCases := []struct {
		name        string
		setupEngine func() *AlluxioEngine
		wantErr     bool
		errContains string
	}{
		{
			name: "emptyOwnerDatasetUID",
			setupEngine: func() *AlluxioEngine {
				engine := mockAlluxioEngineForTests(dataset, alluxioruntime)
				runtimeInfo, _ := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				engine.runtimeInfo = runtimeInfo
				mockedObjects := mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
				resources := []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
				}
				client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
				engine.Client = client
				return engine
			},
			wantErr:     true,
			errContains: "OwnerDatasetUID is not set",
		},
		{
			name: "placementModeNotSet",
			setupEngine: func() *AlluxioEngine {
				engine := mockAlluxioEngineForTests(dataset, alluxioruntime)
				runtimeInfo, _ := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				runtimeInfo.SetOwnerDatasetUID("test-uid")
				engine.runtimeInfo = runtimeInfo
				mockedObjects := mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
				resources := []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
				}
				client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
				engine.Client = client
				return engine
			},
			wantErr:     true,
			errContains: "exclusive mode is not set",
		},
		{
			name: "fullyConfiguredRuntimeInfo",
			setupEngine: func() *AlluxioEngine {
				engine := mockAlluxioEngineForTests(dataset, alluxioruntime)
				runtimeInfo, _ := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				runtimeInfo.SetOwnerDatasetUID("test-uid")
				runtimeInfo.SetupWithDataset(dataset)
				engine.runtimeInfo = runtimeInfo
				mockedObjects := mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
				resources := []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
				}
				client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
				engine.Client = client
				return engine
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := tc.setupEngine()
			err := engine.Validate(ctx)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got %v", err)
				}
			}
		})
	}
}
