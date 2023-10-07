/*
  Copyright 2023 The Fluid Authors.

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

package utils

import (
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTimeleft(t *testing.T) {
	ttl := int32(10)
	condtionTime := v1.NewTime(time.Now())
	testcase := []struct {
		name              string
		dataload          datav1alpha1.DataLoad
		dataoperationType datav1alpha1.OperationType
		validRemaining    bool
		wantErr           bool
	}{
		{
			name: "get remaining time successfully",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							Type:               common.Complete,
							LastProbeTime:      condtionTime,
							LastTransitionTime: condtionTime,
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    true,
			wantErr:           false,
		},
		{
			name: "not set ttl",
			dataload: datav1alpha1.DataLoad{
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							Type:          common.Complete,
							LastProbeTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
		{
			name: "data operation not completion",
			dataload: datav1alpha1.DataLoad{
				Status: datav1alpha1.OperationStatus{},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
		{
			name: "get remaining time < 0",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							Type:          common.Complete,
							LastProbeTime: v1.NewTime(time.Now().Add(-20 * time.Second)),
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
	}
	for _, test := range testcase {
		remaining, err := Timeleft(&test.dataload, &test.dataload.Status, test.dataoperationType)
		if test.validRemaining != (remaining != nil && *remaining > 0) {
			t.Errorf("GetRemaining want validRemaining %v, get remaining %v", test.validRemaining, remaining)
		}
		if test.wantErr != (err != nil) {
			t.Errorf("GetRemaining want error %v, get error %v", test.wantErr, err)
		}
	}
}

func TestGetTTL(t *testing.T) {
	ttl := int32(10)
	testcase := []struct {
		name              string
		dataload          datav1alpha1.DataLoad
		dataoperationType datav1alpha1.OperationType
		ttl               *int32
		wantErr           bool
	}{
		{
			name: "get ttl",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			ttl:               &ttl,
			wantErr:           false,
		},
		{
			name:              "no ttl",
			dataload:          datav1alpha1.DataLoad{},
			dataoperationType: datav1alpha1.DataLoadType,
			ttl:               nil,
			wantErr:           false,
		},
		{
			name:              "wrong data operation type",
			dataload:          datav1alpha1.DataLoad{},
			dataoperationType: datav1alpha1.DataMigrateType,
			ttl:               nil,
			wantErr:           true,
		},
	}
	for _, test := range testcase {
		ttl, err := GetTTL(&test.dataload, test.dataoperationType)
		if ttl != test.ttl {
			t.Errorf("Get wrong ttl value, want %v, get %v", test.ttl, ttl)
		}
		if test.wantErr != (err != nil) {
			t.Errorf("GetTTL want error %v, get error %v", test.wantErr, err)
		}
	}
}

func TestNeedCleanUp(t *testing.T) {
	ttl := int32(10)
	testcase := []struct {
		name        string
		dataload    datav1alpha1.DataLoad
		needCleanUp bool
	}{
		{
			name: "need clean up",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							Type:          common.Complete,
							LastProbeTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			needCleanUp: true,
		},
		{
			name: "have no job condition",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{},
			},
			needCleanUp: false,
		},
		{
			name: "have no ttl",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							Type:          common.Complete,
							LastProbeTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			needCleanUp: false,
		},
	}
	for _, test := range testcase {
		needCleanUp := NeedCleanUp(&test.dataload, &test.dataload.Status, datav1alpha1.DataLoadType)
		if needCleanUp != test.needCleanUp {
			t.Errorf("NeedCleanUp want %v, get %v", test.needCleanUp, needCleanUp)
		}
	}
}
