/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
