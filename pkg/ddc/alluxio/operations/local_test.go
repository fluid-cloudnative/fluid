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

package operations

import (
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestSyncLocalDir(t *testing.T) {
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))
	var tests = []struct {
		path string
		err  error
	}{
		{"/underFSStorage/test", nil},
	}

	for _, test := range tests {
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.SyncLocalDir(test.path)
		// fmt.Println(expectedErr)
		if err == nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}

func TestSoftLink(t *testing.T) {
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))
	var testcases = []struct {
		target   string
		linkName string
		err      error
	}{
		{"/pvcs", "/underFSStorage/test", nil},
	}
	for _, tc := range testcases {
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.SoftLink(tc.target, tc.linkName)
		if err == nil {
			t.Errorf("not failed to create soft link %v to %v", tc.target, tc.linkName)
		}
	}
}
