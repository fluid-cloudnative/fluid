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

package operations

import (
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func mockExecCommandInContainerForSyncLocalDir() (stdout string, stderr string, err error) {
	r := `File Size     In GooseFS       Path
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
		tools := NewGooseFSFileUtils("", "", "", ctrl.Log)
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
