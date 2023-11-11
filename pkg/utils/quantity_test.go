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

package utils

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTranformQuantityToUnits(t *testing.T) {
	testQuantity1 := resource.MustParse("10Gi")
	testQuantity2 := resource.MustParse("3.5Mi")
	testQuantity3 := resource.MustParse("1024Ki")

	tests := []struct {
		name      string
		quantity  *resource.Quantity
		wantValue string
	}{
		{
			name:      "test1 for TransformQuantityToUnits",
			quantity:  &testQuantity1,
			wantValue: "10GiB",
		},
		{
			name:      "test2 for TransformQuantityToUnits",
			quantity:  &testQuantity2,
			wantValue: "3.5MiB",
		},
		{
			name:      "test3 for TransformQuantityToUnits",
			quantity:  &testQuantity3,
			wantValue: "1MiB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := TranformQuantityToUnits(tt.quantity); gotValue != tt.wantValue {
				t.Errorf("TranformQuantityToUnits() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTransformQuantityToAlluxioUnit(t *testing.T) {
	testQuantity1 := resource.MustParse("10Gi")
	testQuantity2 := resource.MustParse("10M")

	tests := []struct {
		name      string
		quantity  *resource.Quantity
		wantValue string
	}{
		{
			name:      "test1 for TransformQuantityToAlluxioUnit",
			quantity:  &testQuantity1,
			wantValue: "10GB",
		},
		{
			name:      "test2 for TransformQuantityToAlluxioUnit",
			quantity:  &testQuantity2,
			wantValue: "10M",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := TransformQuantityToAlluxioUnit(tt.quantity); gotValue != tt.wantValue {
				t.Errorf("TransformQuantityToAlluxioUnit() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTransformQuantityToJindoUnit(t *testing.T) {
	testQuantity1 := resource.MustParse("5Gi")

	tests := []struct {
		name      string
		quantity  *resource.Quantity
		wantValue string
	}{
		{
			name:      "test1 for TransformQuantityToJindoUnit",
			quantity:  &testQuantity1,
			wantValue: "5g",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := TransformQuantityToJindoUnit(tt.quantity); gotValue != tt.wantValue {
				t.Errorf("TransformQuantityToJindoUnit() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTransformQuantityToGooseFSUnit(t *testing.T) {
	testQuantity1 := resource.MustParse("10Gi")
	testQuantity2 := resource.MustParse("10M")

	tests := []struct {
		name      string
		quantity  *resource.Quantity
		wantValue string
	}{
		{
			name:      "test1 for TransformQuantityToGooseFSUnit",
			quantity:  &testQuantity1,
			wantValue: "10GB",
		},
		{
			name:      "test2 for TransformQuantityToGooseFSUnit",
			quantity:  &testQuantity2,
			wantValue: "10M",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := TransformQuantityToGooseFSUnit(tt.quantity); gotValue != tt.wantValue {
				t.Errorf("TransformQuantityToGooseFSUnit() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTransformQuantityToEFCUnit(t *testing.T) {
	testQuantity1 := resource.MustParse("10Gi")
	testQuantity2 := resource.MustParse("10M")

	tests := []struct {
		name      string
		quantity  *resource.Quantity
		wantValue string
	}{
		{
			name:      "test1 for TransformQuantityToEFCUnit",
			quantity:  &testQuantity1,
			wantValue: "10GB",
		},
		{
			name:      "test2 for TransformQuantityToEFCUnit",
			quantity:  &testQuantity2,
			wantValue: "10M",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := TransformQuantityToEFCUnit(tt.quantity); gotValue != tt.wantValue {
				t.Errorf("TransformQuantityToEFCUnit() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}
