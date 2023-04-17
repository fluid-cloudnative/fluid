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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"
)

type EACFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewEACFileUtils(podName string, containerName string, namespace string, log logr.Logger) EACFileUtils {
	return EACFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (a EACFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = a.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		a.log.Info("execute in time", "command", securityutil.FilterCommand(command))
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}

	return
}

// execWithoutTimeout
func (a EACFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
	stdout, stderr, err = kubeclient.ExecCommandInContainer(a.podName, a.container, a.namespace, command)
	if err != nil {
		a.log.Info("Stdout", "Command", command, "Stdout", stdout)
		a.log.Error(err, "Failed", "Command", command, "FailedReason", stderr)
		return
	}
	if verbose {
		a.log.Info("Stdout", "Command", command, "Stdout", stdout)
	}

	return
}

func (a EACFileUtils) IsExist(eacSubPath string) (found bool, err error) {
	var (
		command = []string{"ls", eacSubPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		if strings.Contains(stdout, "No such file or directory") || strings.Contains(stderr, "No such file or directory") {
			return false, nil
		} else {
			err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
			return false, err
		}
	} else {
		found = true
	}
	return
}

func (a EACFileUtils) DeleteDir(dir string) (err error) {
	var (
		command = []string{"rm", "-rf", dir}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	return
}

func (a EACFileUtils) Ready() (ready bool) {
	var (
		command = []string{"mount", "|", "grep", common.EFCMountType}
	)

	_, _, err := a.exec(command, true)
	if err == nil {
		ready = true
	}

	return ready
}
