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

	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
)

type JindoFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewJindoFileUtils(podName string, containerName string, namespace string, log logr.Logger) JindoFileUtils {

	return JindoFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (a JindoFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = a.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		a.log.V(1).Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}

	return
}

// exec with timeout
func (a JindoFileUtils) execWithTimeOut(command []string, verbose bool, second int64) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*time.Duration(second))
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = a.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		a.log.V(1).Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}

	return
}

// execWithoutTimeout
func (a JindoFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
	err = cmdguard.ValidateCommandSlice(command)
	if err != nil {
		return
	}

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

// Get summary info of the jindo Engine
func (a JindoFileUtils) ReportSummary() (summary string, err error) {
	var (
		command = []string{"jindo", "fs", "-report"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return stdout, err
	}
	return stdout, err
}

func (a JindoFileUtils) IsMounted(mountPoint string) (mounted bool, err error) {
	var (
		command = []string{"jindo", "admin", "-mount"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		return mounted, fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
	}

	results := strings.Split(stdout, "\n")

	for _, line := range results {
		fields := strings.Fields(line)
		if len(fields) > 2 && fields[2] == mountPoint {
			mounted = true
			return mounted, nil
		}
	}

	return mounted, err
}

func (a JindoFileUtils) Mount(mountName string, ufsPath string) (err error) {

	var (
		command = []string{"jindo", "admin", "-mount"}
	)
	// jindo fsxadmin -mount /path oss://xyz/
	if strings.HasPrefix(mountName, "/") {
		command = append(command, mountName, ufsPath)
	} else {
		command = append(command, "/"+mountName, ufsPath)
	}

	_, _, _ = a.exec(command, false)
	/*if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}*/

	return nil
}

func (a JindoFileUtils) GetUfsTotalSize(url string) (summary string, err error) {
	var (
		command = []string{"hadoop", "fs", "-count", url}
		stdout  string
		stderr  string
	)

	// default 2min
	stdout, stderr, err = a.execWithTimeOut(command, false, 120)

	str := strings.Fields(stdout)

	if len(str) < 3 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	stdout = str[2]

	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return stdout, err
	}
	return stdout, err
}

// Check if the JIndo is ready by running `jindo jfs -report` command
func (a JindoFileUtils) Ready() (ready bool) {
	var (
		command = []string{"jindo", "fs", "-report"}
	)

	_, _, err := a.exec(command, true)
	if err == nil {
		ready = true
	}

	return ready
}

// IsExist checks if the JindoPath exists
func (a JindoFileUtils) IsExist(jindoPath string) (found bool, err error) {
	var (
		command = []string{"hadoop", "fs", "-ls", "jindo://" + jindoPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		if strings.Contains(stdout, "No such file or directory") {
			err = nil
		} else {
			err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
			return
		}
	} else {
		found = true
	}
	return
}
