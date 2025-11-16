/*
Copyright 2021 The Fluid Authors.
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
package operations

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestJindoFIlUtils_CleanCache(t *testing.T) {
	ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stout", "", nil
	}
	ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(JindoFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := &JindoFileUtils{log: fake.NullLogger()}
	err := a.CleanCache()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(JindoFileUtils{}), "exec", ExecCommon)
	err = a.CleanCache()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}