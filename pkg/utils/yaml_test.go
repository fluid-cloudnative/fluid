
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
