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

package operations

import (
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutils "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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

// Get summary info of the jindo Engine
func (a JindoFileUtils) ReportSummary() (summary string, err error) {
	var (
		command = []string{"jindocache", "-report"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.ReportSummary() failed", "stdout", stdout, "stderr", stderr)
		return stdout, err
	}
	return stdout, err
}

func (a JindoFileUtils) IsMounted(mountPoint string) (mounted bool, err error) {
	var (
		command = []string{"jindocache", "-mount"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.IsMounted() failed", "stdout", stdout, "stderr", stderr)
		return mounted, err
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

func (a JindoFileUtils) Mount(mountPathInJindo string, ufsPath string) (err error) {

	var (
		command = []string{"jindocache", "-mount", mountPathInJindo, ufsPath}
		stdout  string
		stderr  string
	)
	// jindo fsxadmin -mount /path oss://xyz/

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		if strings.Contains(stdout, "Mount point already exists") {
			// ignore existed mount points
			err = nil
			return
		}
		a.log.Error(err, "JindoFileUtils.Mount() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	return nil
}

func (a JindoFileUtils) GetUfsTotalSize(url string) (summary string, err error) {
	var (
		command = []string{"jindo", "fs", "-count", url}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.GetUfsTotalSize() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	str := strings.Fields(stdout)
	if len(str) < 3 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	summary = str[2]
	return
}

// Check if the Jindo is ready by running `jindocache -report` command
func (a JindoFileUtils) Ready() (ready bool) {
	var (
		command = []string{"jindocache", "-report"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err := a.exec(command, true)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.Ready() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	ready = true
	return
}

func (a JindoFileUtils) IsRefreshed() (refreshed bool, err error) {
	var (
		command = []string{"jindocache", "-listCacheSet"}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.IsRefreshed() failed", "stdout", stdout, "stderr", stderr)
		return refreshed, err
	}
	results := strings.Split(stdout, "\n")
	for _, line := range results {
		if strings.Contains(line, "cacheStrategy") {
			refreshed = true
		}
	}
	return refreshed, err
}

func (a JindoFileUtils) RefreshCacheSet() (err error) {
	var (
		command = []string{"jindocache", "-refreshCacheSet", "-path", "/cacheset.xml"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		a.log.Error(err, "JindoFileUtils.RefreshCacheSet() failed", "stdout", stdout, "stderr", stderr)
		return
	}
	return
}
