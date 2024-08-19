/*
  Copyright 2022 The Fluid Authors.

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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutils "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type EFCFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewEFCFileUtils(podName string, containerName string, namespace string, log logr.Logger) EFCFileUtils {
	return EFCFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (a EFCFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	// redact sensitive info in command for printing
	redactedCommand := securityutils.FilterCommand(command)

	stdout, stderr, err = kubeclient.ExecCommandInContainerWithTimeout(a.podName, a.container, a.namespace, command, common.FileUtilsExecTimeout)
	if err != nil {
		err = errors.Wrapf(err, "error when executing command %v", redactedCommand)
		return
	}

	if verbose {
		a.log.Info("Exec command succeeded", "command", redactedCommand, "stdout", stdout, "stderr", stderr)
	}

	return
}

func (a EFCFileUtils) DeleteDir(dir string) (err error) {
	var (
		command = []string{"rm", "-rf", dir}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		a.log.Error(err, "EFCFileUtils.DeleteDir() failed", "stdout", stdout, "stderr", stderr)
		return
	}
	return
}

func (a EFCFileUtils) Ready() (ready bool) {
	var (
		command = []string{"mount", "|", "grep", common.EFCMountType}
		stdout  string
		stderr  string
	)

	stdout, stderr, err := a.exec(command, true)
	if err != nil {
		a.log.Error(err, "EFCFileUtils.Ready() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	ready = true
	return
}
