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
	"reflect"
	"testing"
)

func TestFillSliceWithString(t *testing.T) {
	type args struct {
		str string
		num int
	}
	tests := []struct {
		name string
		args args
		want *[]string
	}{
		{
			name: "Fill Slice Test1",
			args: args{
				str: "foo",
				num: 3,
			},
			want: &[]string{"foo", "foo", "foo"},
		},
		{
			name: "Fill Slice Test2",
			args: args{
				str: "bar",
				num: 0,
			},
			want: &[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FillSliceWithString(tt.args.str, tt.args.num); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FillSliceWithString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubtractString(t *testing.T) {
	testcases := map[string]struct {
		firstStr []string
		secStr   []string
		want     []string
	}{
		"SubtractString test case 1": {
			firstStr: []string{"foo", "bar", "hello", "bar"},
			secStr:   []string{"foo", "bar"},
			want:     []string{"hello"},
		},
		"SubtractString test case 2": {
			firstStr: []string{"foo", "bar", "hello", "bar", "world", "world"},
			secStr:   []string{"foo", "bar"},
			want:     []string{"hello", "world", "world"},
		},
		"SubtractString test case 3": {
			firstStr: []string{"bar", "bar"},
			secStr:   []string{},
			want:     []string{"bar", "bar"},
		},
		"SubtractString test case 4": {
			firstStr: []string{},
			secStr:   []string{"bar", "bar"},
			want:     []string{},
		},
	}

	for k, item := range testcases {
		result := SubtractString(item.firstStr, item.secStr)
		if !reflect.DeepEqual(item.want, result) {
			t.Errorf("%s check failure,want:%v,got:%v", k, item.want, result)
		}
	}

}

func TestRemoveDuplicateStr(t *testing.T) {

	input := []string{"Mumbai", "Delhi", "Ahmedabad", "Mumbai", "Bangalore", "Delhi", "Kolkata", "Pune"}
	expected := []string{"Mumbai", "Delhi", "Ahmedabad", "Bangalore", "Kolkata", "Pune"}

	result := RemoveDuplicateStr(input)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("check RemoveDuplicateStr failure,want:%v,got:%v", expected, result)
	}
}
