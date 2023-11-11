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
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"

	"gopkg.in/yaml.v2"
)

type dummyOuter struct {
	Foo   string     `yaml:"foo"`
	Bar   int        `yaml:"bar"`
	Inner dummyInner `yaml:"inner"`
}

type dummyInner struct {
	Name      string            `yaml:"name"`
	KeyValues map[string]string `yaml:"keyValues,omitempty"`
}

func TestToYaml(t *testing.T) {
	tempFile, err := os.CreateTemp(os.TempDir(), "dummy")
	if err != nil {
		t.Errorf("TestToYaml can't write temp file. error = %v", err)
		t.FailNow()
	}

	dummy := dummyOuter{
		Foo: "foo",
		Bar: rand.Int(),
		Inner: dummyInner{
			Name: strconv.Itoa(rand.Int()),
			KeyValues: map[string]string{
				"foo": "bar",
				"xxx": "yyy",
			},
		},
	}

	err = ToYaml(dummy, tempFile)
	if err != nil {
		t.Errorf("ToYaml() error = %v, expected error = nil", err)
		t.FailNow()
	}

	tempFileName := tempFile.Name()
	bytes, err := os.ReadFile(tempFileName)
	if err != nil {
		t.Errorf("os. ReadFile() error = %v, expected error = nil", err)
		t.FailNow()
	}

	var deserializedDummy dummyOuter
	err = yaml.Unmarshal(bytes, &deserializedDummy)
	if err != nil {
		t.Errorf("yaml.Unmarshal() error = %v, expected error = nil", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(deserializedDummy, dummy) {
		t.Errorf("Expected got %v, but got %v", dummy, deserializedDummy)
	}
}
