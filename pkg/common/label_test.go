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
package common

import (
	"reflect"
	"testing"
)

func TestLabelToModify(t *testing.T) {
	var testCase = []struct {
		labelToModify              LabelToModify
		expectedLabelKey           string
		expectedLabelValue         string
		expectedLabelOperationType OperationType
	}{
		{
			labelToModify: LabelToModify{
				labelKey:      "commonLabel",
				labelValue:    "true",
				operationType: AddLabel,
			},
			expectedLabelKey:           "commonLabel",
			expectedLabelValue:         "true",
			expectedLabelOperationType: AddLabel,
		},
	}

	for _, test := range testCase {
		labelKey := test.labelToModify.GetLabelKey()
		labelValue := test.labelToModify.GetLabelValue()
		operationType := test.labelToModify.GetOperationType()

		if labelKey != test.expectedLabelKey || labelValue != test.expectedLabelValue || operationType != test.expectedLabelOperationType {
			t.Errorf("expected labelKey %s, get labelKey %s; expected labelValue %s, get labelValue %s; expected labelOperationType %s, get labelOperationType %s",
				test.expectedLabelKey, labelKey, test.expectedLabelValue, labelValue, test.expectedLabelOperationType, operationType)
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
					labelKey:      "commonLabel",
					labelValue:    "true",
					operationType: AddLabel,
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

func TestOperator(t *testing.T) {
	var labelsToModify LabelsToModify

	var testCase = []struct {
		labelKey      string
		labelValue    string
		operationType OperationType
		expectedLabel LabelToModify
	}{
		{
			labelKey:      "commonLabel",
			labelValue:    "true",
			operationType: AddLabel,
			expectedLabel: LabelToModify{
				labelKey:      "commonLabel",
				labelValue:    "true",
				operationType: AddLabel,
			},
		},
		{
			labelKey:      "commonLabel",
			labelValue:    "true",
			operationType: DeleteLabel,
			expectedLabel: LabelToModify{
				labelKey:      "commonLabel",
				operationType: DeleteLabel,
			},
		},
	}

	for index, test := range testCase {
		labelsToModify.operator(test.labelKey, test.labelValue, test.operationType)
		if !reflect.DeepEqual(labelsToModify.labels[index], test.expectedLabel) {
			t.Errorf("fail to add the label")
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
					labelKey:      "commonLabel",
					labelValue:    "true",
					operationType: AddLabel,
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
					labelKey:      "commonLabel",
					operationType: DeleteLabel,
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
					labelKey:      "commonLabel",
					operationType: UpdateLabel,
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

func TestHitTarget(t *testing.T) {
	testCases := map[string]struct {
		labels      map[string]string
		target      string
		targetValue string
		wantHit     bool
	}{
		"test label target hit case 1": {
			labels:      map[string]string{EnableFluidInjectionFlag: "true"},
			target:      EnableFluidInjectionFlag,
			targetValue: "true",
			wantHit:     true,
		},
		"test label target hit case 2": {
			labels:      map[string]string{EnableFluidInjectionFlag: "false"},
			target:      EnableFluidInjectionFlag,
			targetValue: "true",
			wantHit:     false,
		},
		"test label target hit case 3": {
			labels:      nil,
			target:      EnableFluidInjectionFlag,
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

func TestLabelAnnotationPodSchedRegex(t *testing.T) {
	testCases := map[string]struct {
		target string
		got    string
		match  bool
	}{
		"correct": {
			target: LabelAnnotationDataset + ".dsA.sched",
			match:  true,
			got:    "dsA",
		},
		"wrong fluid.io": {
			target: "fluidaio/dataset.dsA.sched",
			match:  false,
		},
		"wrong prefix": {
			target: "a.fluid.io/dataset.dsA.sched",
			match:  false,
		},
	}

	for index, item := range testCases {
		submatch := LabelAnnotationPodSchedRegex.FindStringSubmatch(item.target)

		if !item.match && len(submatch) == 2 {
			t.Errorf("[%s] check match, want:%t, got:%t", index, item.match, len(submatch) == 2)
		}

		if len(submatch) == 2 && submatch[1] != item.got {
			t.Errorf("[%s] check failure, want:%s, got:%s", index, item.got, submatch[1])
		}
	}
}
