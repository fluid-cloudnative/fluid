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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutils "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type ThinFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewThinFileUtils(podName string, containerName string, namespace string, log logr.Logger) ThinFileUtils {
	return ThinFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (t ThinFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	// redact sensitive info in command for printing
	redactedCommand := securityutils.FilterCommand(command)

	t.log.V(1).Info("Exec command start", "command", redactedCommand)
	stdout, stderr, err = kubeclient.ExecCommandInContainerWithTimeout(t.podName, t.container, t.namespace, command, common.FileUtilsExecTimeout)
	if err != nil {
		err = errors.Wrapf(err, "error when executing command %v", redactedCommand)
		return
	}
	t.log.V(1).Info("Exec command finished", "command", redactedCommand)

	if verbose {
		t.log.Info("Exec command succeeded", "command", redactedCommand, "stdout", stdout, "stderr", stderr)
	}

	return
}

// Load the metadata without timeout
func (t ThinFileUtils) LoadMetadataWithoutTimeout(path string) (err error) {
	var (
		command = []string{"ls", "-lR", path}
		stdout  string
		stderr  string
	)

	start := time.Now()
	stdout, stderr, err = t.exec(command, false)
	duration := time.Since(start)
	t.log.Info("Async Load Metadata took times to run", "period", duration)
	if err != nil {
		t.log.Error(err, "ThinFileUtils.LoadMetadataWithoutTimeout() failed", "stdout", stdout, "stderr", stderr)
		return
	}
	t.log.Info("Async Load Metadata finished")
	return
}

// GetUsedSpace Get used space in byte
// use "df --block-size=1 |grep <path>'"
func (t ThinFileUtils) GetUsedSpace(path string) (usedSpace int64, err error) {
	var (
		strs    = fmt.Sprintf(`df --block-size=1 |grep %s`, path)
		command = []string{"bash", "-c", strs}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = t.exec(command, false)
	if err != nil {
		t.log.Error(err, "ThinFileUtils.GetUsedSpace() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	// [<Filesystem>       <Size>  <Used> <Avail> <Use>% <Mounted on>]
	str := strings.TrimSuffix(stdout, "\n")

	data := strings.Fields(str)
	if len(data) < 6 || len(data)%6 != 0 {
		err = fmt.Errorf("failed to parse %s in GetUsedSpace method", data)
		return
	}

	for i := 0; i < len(data); i += 6 {
		used, err := strconv.ParseInt(data[2+i], 10, 64)
		if err != nil {
			return usedSpace, err
		}
		usedSpace += used
	}

	return usedSpace, err
}

func (t ThinFileUtils) GetFileCount(path string) (fileCount int64, err error) {
	var (
		strs    = fmt.Sprintf("ls -lR %s |grep ^- |wc -l ", path)
		command = []string{"bash", "-c", strs}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = t.exec(command, false)
	if err != nil {
		t.log.Error(err, "ThinFileUtils.GetFileCount() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	// eg: Master.FilesCompleted  (Type: COUNTER, Value: 6,367,897)
	str := strings.Split(stdout, "\n")

	if len(str) != 1 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	data := strings.Fields(str[0])
	if len(data) != 1 {
		err = fmt.Errorf("failed to parse %s in Count method", data)
		return
	}

	fileCount, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return
	}

	return fileCount, nil
}
