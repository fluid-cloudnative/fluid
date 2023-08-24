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
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
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

// Load the metadata without timeout
func (t ThinFileUtils) LoadMetadataWithoutTimeout(path string) (err error) {
	var (
		command = []string{"ls", "-lR", path}
		stdout  string
		stderr  string
	)

	start := time.Now()
	stdout, stderr, err = t.execWithoutTimeout(command, false)
	duration := time.Since(start)
	t.log.Info("Async Load Metadata took times to run", "period", duration)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	} else {
		t.log.Info("Async Load Metadata finished", "stdout", stdout)
	}
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
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	// [<Filesystem>       <Size>  <Used> <Avail> <Use>% <Mounted on>]
	str := strings.TrimSuffix(stdout, "\n")

	data := strings.Fields(str)
	if len(data) < 6 {
		err = fmt.Errorf("failed to parse %s in GetUsedSpace method", data)
		return
	}

	usedSpace, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return
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
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
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

// exec with timeout
func (t ThinFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = t.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		t.log.Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}
	return
}

// execWithoutTimeout
func (t ThinFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
	// validate the pipe command with white list
	err = cmdguard.ValidateCommandSlice(command)
	if err != nil {
		return
	}

	stdout, stderr, err = kubeclient.ExecCommandInContainer(t.podName, t.container, t.namespace, command)
	if err != nil {
		t.log.Info("Stdout", "Command", command, "Stdout", stdout)
		t.log.Error(err, "Failed", "Command", command, "FailedReason", stderr)
		return
	}
	if verbose {
		t.log.Info("Stdout", "Command", command, "Stdout", stdout)
	}

	return
}
