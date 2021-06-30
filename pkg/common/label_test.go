/*

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
package common

import (
	"reflect"
	"testing"
)

func TestHitTarget(t *testing.T) {
	testCases := map[string]struct {
		labels      map[string]string
		target      string
		targetValue string
		wantHit     bool
	}{
		"test label target hit case 1": {
			labels:      map[string]string{LabelFluidSchedulingStrategyFlag: "true"},
			target:      LabelFluidSchedulingStrategyFlag,
			targetValue: "true",
			wantHit:     true,
		},
		"test label target hit case 2": {
			labels:      map[string]string{LabelFluidSchedulingStrategyFlag: "false"},
			target:      LabelFluidSchedulingStrategyFlag,
			targetValue: "true",
			wantHit:     false,
		},
		"test label target hit case 3": {
			labels:      nil,
			target:      LabelFluidSchedulingStrategyFlag,
			targetValue: "true",
			wantHit:     false,
		},
	}

	for index, item := range testCases {
		gotHit := CheckExpectValue(item.labels, item.target, item.targetValue)
		if gotHit != item.wantHit {
			t.Errorf("%s check failure, want:%t,got:%t", index, item.wantHit, gotHit)
		}
	}

}

func TestGetLabels(t *testing.T) {
	var testCase = []struct {
		labelsToModify LabelsToModify
		expectedResult []LabelToModify
	}{
		{
			labelsToModify: LabelsToModify{},
			expectedResult: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: AddLabel,
				},
			},
		},
	}

	for _, test := range testCase {
		test.labelsToModify.Add("commonLabel", "true")
		label := test.labelsToModify.GetLabels()
		if !reflect.DeepEqual(label, test.expectedResult) {
			t.Errorf("fail to get the labels")
		}
	}

}

func TestAdd(t *testing.T) {
	var testCase = []struct {
		labelKey             string
		labelValue           string
		wantedLabelsToModify []LabelToModify
	}{
		{
			labelKey:   "commonLabel",
			labelValue: "true",
			wantedLabelsToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: AddLabel,
				},
			},
		},
	}

	var labelsToModify LabelsToModify
	for _, test := range testCase {
		labelsToModify.Add(test.labelKey, test.labelValue)
		if !reflect.DeepEqual(labelsToModify.GetLabels(), test.wantedLabelsToModify) {
			t.Errorf("fail to add labe to modify to slice")
		}
	}
}

func TestDelete(t *testing.T) {
	var testCase = []struct {
		labelKey             string
		labelValue           string
		wantedLabelsToModify []LabelToModify
	}{
		{
			labelKey: "commonLabel",
			wantedLabelsToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					OperationType: DeleteLabel,
				},
			},
		},
	}

	var labelsToModify LabelsToModify
	for _, test := range testCase {
		labelsToModify.Delete(test.labelKey)
		if !reflect.DeepEqual(labelsToModify.GetLabels(), test.wantedLabelsToModify) {
			t.Errorf("fail to add labe to modify to slice")
		}
	}
}

func TestUpdate(t *testing.T) {
	var testCase = []struct {
		labelKey             string
		labelValue           string
		wantedLabelsToModify []LabelToModify
	}{
		{
			labelKey: "commonLabel",
			wantedLabelsToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					OperationType: UpdateLabel,
				},
			},
		},
	}

	var labelsToModify LabelsToModify
	for _, test := range testCase {
		labelsToModify.Update(test.labelKey, test.labelValue)
		if !reflect.DeepEqual(labelsToModify.GetLabels(), test.wantedLabelsToModify) {
			t.Errorf("fail to add labe to modify to slice")
		}
	}
}
