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
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
)

type AlluxioFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewAlluxioFileUtils(podName string, containerName string, namespace string, log logr.Logger) AlluxioFileUtils {

	return AlluxioFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// IsExist checks if the alluxioPath exists
func (a AlluxioFileUtils) IsExist(alluxioPath string) (found bool, err error) {
	var (
		command = []string{"alluxio", "fs", "ls", alluxioPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		if strings.Contains(stdout, "does not exist") {
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

// Get summary info of the Alluxio Engine
func (a AlluxioFileUtils) ReportSummary() (summary string, err error) {
	var (
		command = []string{"alluxio", "fsadmin", "report", "summary"}
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

// Load the metadata without timeout
func (a AlluxioFileUtils) LoadMetadataWithoutTimeout(alluxioPath string) (err error) {
	var (
		command = []string{"alluxio", "fs", "loadMetadata", "-R", alluxioPath}
		stdout  string
		stderr  string
	)

	start := time.Now()
	stdout, stderr, err = a.execWithoutTimeout(command, false)
	duration := time.Since(start)
	a.log.Info("Async Load Metadata took times to run", "period", duration)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	} else {
		a.log.Info("Async Load Metadata finished", "stdout", stdout)
	}
	return
}

// LoadMetaData loads the metadata.
func (a AlluxioFileUtils) LoadMetaData(alluxioPath string, sync bool) (err error) {
	var (
		// command = []string{"alluxio", "fs", "-Dalluxio.user.file.metadata.sync.interval=0", "ls", "-R", alluxioPath}
		// command = []string{"alluxio", "fs", "-Dalluxio.user.file.metadata.sync.interval=0", "count", alluxioPath}
		command []string
		stdout  string
		stderr  string
	)

	if sync {
		command = []string{"alluxio", "fs", "-Dalluxio.user.file.metadata.sync.interval=0", "ls", "-R", alluxioPath}
	} else {
		command = []string{"alluxio", "fs", "ls", "-R", alluxioPath}
	}

	start := time.Now()
	stdout, stderr, err = a.exec(command, false)
	duration := time.Since(start)
	a.log.Info("Load MetaData took times to run", "period", duration)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}

func (a AlluxioFileUtils) Mkdir(alluxioPath string) (err error) {
	var (
		command = []string{"alluxio", "fs", "mkdir", alluxioPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}

func (a AlluxioFileUtils) Mount(alluxioPath string,
	ufsPath string,
	options map[string]string,
	readOnly bool,
	shared bool) (err error) {

	// exist, expectedErr := a.IsExist(alluxioPath)
	// if expectedErr != nil {
	// 	return expectedErr
	// }

	// if !exist {
	// 	expectedErr = a.Mkdir(alluxioPath)
	// 	if expectedErr != nil {
	// 		return expectedErr
	// 	}
	// }

	var (
		command = []string{"alluxio", "fs", "mount"}
		stderr  string
		stdout  string
	)

	if readOnly {
		command = append(command, "--readonly")
	}

	if shared {
		command = append(command, "--shared")
	}

	for key, value := range options {
		command = append(command, "--option", fmt.Sprintf("%s=%s", key, value))
	}

	command = append(command, alluxioPath, ufsPath)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}

func (a AlluxioFileUtils) IsMounted(alluxioPath string) (mounted bool, err error) {
	var (
		command = []string{"alluxio", "fs", "mount"}
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
		a.log.Info("parse output of isMounted", "alluxioPath", alluxioPath, "fields", fields)
		if fields[2] == alluxioPath {
			mounted = true
			return mounted, nil
		}
	}

	// pattern := fmt.Sprintf(" on %s ", alluxioPath)
	// if strings.Contains(stdout, pattern) {
	// 	mounted = true
	// }

	return mounted, err
}

// Check if the Alluxio is ready by running `alluxio fsadmin report` command
func (a AlluxioFileUtils) Ready() (ready bool) {
	var (
		command = []string{"alluxio", "fsadmin", "report"}
	)

	_, _, err := a.exec(command, true)
	if err == nil {
		ready = true
	}

	return ready
}

func (a AlluxioFileUtils) Du(alluxioPath string) (ufs int64, cached int64, cachedPercentage string, err error) {
	var (
		command = []string{"alluxio", "fs", "du", "-s", alluxioPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	str := strings.Split(stdout, "\n")

	if len(str) != 2 {
		err = fmt.Errorf("failed to parse %s in Du method", str)
		return
	}

	data := strings.Fields(str[1])
	if len(data) != 4 {
		err = fmt.Errorf("failed to parse %s in Du method", data)
		return
	}

	ufs, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return
	}

	cached, err = strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return
	}

	cachedPercentage = strings.TrimLeft(data[2], "(")
	cachedPercentage = strings.TrimRight(cachedPercentage, ")")

	return
}

// The count of the Alluxio Filesystem
func (a AlluxioFileUtils) Count(alluxioPath string) (fileCount int64, folderCount int64, total int64, err error) {
	var (
		command                          = []string{"alluxio", "fs", "count", alluxioPath}
		stdout                           string
		stderr                           string
		ufileCount, ufolderCount, utotal uint64
	)

	stdout, stderr, err = a.execWithoutTimeout(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	// [File Count Folder Count Total Bytes 1152 4 154262709011]
	str := strings.Split(stdout, "\n")

	if len(str) != 2 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	data := strings.Fields(str[1])
	if len(data) != 3 {
		err = fmt.Errorf("failed to parse %s in Count method", data)
		return
	}

	ufileCount, err = strconv.ParseUint(data[0], 10, 64)
	if err != nil {
		return
	}

	ufolderCount, err = strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		return
	}

	utotal, err = strconv.ParseUint(data[2], 10, 64)
	if err != nil {
		return
	}

	return int64(ufileCount), int64(ufolderCount), int64(utotal), err
}

// file count of the Alluxio Filesystem (except folder)
// use "alluxio fsadmin report metrics" for better performance
func (a AlluxioFileUtils) GetFileCount() (fileCount int64, err error) {
	args := []string{"alluxio", "fsadmin", "report", "metrics", "|", "grep", "Master.FilesCompleted"}
	var (
		command = []string{"bash", "-c", strings.Join(args, " ")}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.execWithoutTimeout(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	// eg: Master.FilesCompleted  (Type: COUNTER, Value: 6,367,897)
	outStrWithoutComma := strings.Replace(stdout, ",", "", -1)
	matchExp := regexp.MustCompile(`\d+`)
	fileCountStr := matchExp.FindString(outStrWithoutComma)
	fileCount, err = strconv.ParseInt(fileCountStr, 10, 64)
	if err != nil {
		return
	}
	return fileCount, nil
}

// exec with timeout
func (a AlluxioFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
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

// execWithoutTimeout
func (a AlluxioFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
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
