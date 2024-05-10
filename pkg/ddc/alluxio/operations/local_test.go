/*
Copyright 2020 The Fluid Authors.

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

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func mockExecCommandInContainerForSyncLocalDir() (stdout string, stderr string, err error) {
	r := `File Size     In Alluxio       Path
	592.06MB      0B (0%)          /`
	return r, "", nil
}

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
		patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			stdout, stderr, err := mockExecCommandInContainerForSyncLocalDir()
			return stdout, stderr, err
		})
		defer patch1.Reset()
		err := tools.SyncLocalDir(test.path)
		// fmt.Println(expectedErr)
		if err != nil {
			t.Errorf("expected %v, got %v %s", test.path, tools, err)
		}
	}
}
