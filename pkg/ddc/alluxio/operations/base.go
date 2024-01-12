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
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

/*
MetadataInfoFile is a yaml file to save the metadata info of dataset, such as ufs total and fileNum
it is in the form of：
	dataset: <Dataset>
	namespace: <Namespace>
	ufstotal: <ufstotal>
	filenum: <filenum>
*/

type KeyOfMetaDataFile string

var (
	DatasetName KeyOfMetaDataFile = "dataset"
	Namespace   KeyOfMetaDataFile = "namespace"
	UfsTotal    KeyOfMetaDataFile = "ufstotal"
	FileNum     KeyOfMetaDataFile = "filenum"
)

// QueryMetaDataInfoIntoFile queries the metadata info file.
func (a AlluxioFileUtils) QueryMetaDataInfoIntoFile(key KeyOfMetaDataFile, filename string) (value string, err error) {
	line := ""
	switch key {
	case DatasetName:
		line = "1p"
	case Namespace:
		line = "2p"
	case UfsTotal:
		line = "3p"
	case FileNum:
		line = "4p"
	default:
		a.log.Error(errors.New("the key not in  metadatafile"), "key", key)
	}
	var (
		str     = "sed -n '" + line + "' " + filename
		command = []string{"bash", "-c", str}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with  expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
	} else {
		value = strings.TrimPrefix(stdout, string(key)+": ")
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

func (a AlluxioFileUtils) UnMount(alluxioPath string) (err error) {
	var (
		command = []string{"alluxio", "fs", "unmount"}
		stderr  string
		stdout  string
	)

	command = append(command, alluxioPath)

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

func (a AlluxioFileUtils) GetMountedAlluxioPaths() ([]string, error) {
	var (
		command = []string{"alluxio", "fs", "mount"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err := a.exec(command, true)
	if err != nil {
		return []string{}, fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
	}

	results := strings.Split(stdout, "\n")
	var mountedPaths []string
	for _, line := range results {
		fields := strings.Fields(line)
		mountedPaths = append(mountedPaths, fields[2])
	}

	return mountedPaths, err
}

func (a AlluxioFileUtils) FindUnmountedAlluxioPaths(alluxioPaths []string) ([]string, error) {
	mountedPaths, err := a.GetMountedAlluxioPaths()
	if err != nil {
		return []string{}, err
	}

	return utils.SubtractString(alluxioPaths, mountedPaths), err
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
		ufileCount, ufolderCount, utotal int64
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

	ufileCount, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return
	}

	ufolderCount, err = strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return
	}

	utotal, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return
	}

	if ufileCount < 0 || ufolderCount < 0 || utotal < 0 {
		err = fmt.Errorf("the return value of Count method is negative")
		return
	}

	return ufileCount, ufolderCount, utotal, err
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

// ReportMetrics get alluxio metrics by running `alluxio fsadmin report metrics` command
func (a AlluxioFileUtils) ReportMetrics() (metrics string, err error) {
	var (
		command = []string{"alluxio", "fsadmin", "report", "metrics"}
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

// ReportCapacity get alluxio capacity info by running `alluxio fsadmin report capacity` command
func (a AlluxioFileUtils) ReportCapacity() (report string, err error) {
	var (
		command = []string{"alluxio", "fsadmin", "report", "capacity"}
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
		a.log.Info("execute in time", "command", securityutil.FilterCommand(command))
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

func (a AlluxioFileUtils) MasterPodName() (masterPodName string, err error) {
	var (
		command = []string{"alluxio", "fsadmin", "report"}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return a.podName, err
	}

	str := strings.Split(stdout, "\n")
	if len(str) < 1 {
		message := fmt.Sprintf("get wrong result when using command %v", command)
		return a.podName, errors.New(message)
	}

	data := strings.Fields(str[1])
	if len(data) < 2 {
		message := fmt.Sprintf("get wrong result when using command %v", command)
		return a.podName, errors.New(message)
	}
	address := strings.Split(data[2], ":")[0]

	return address, nil
}

func (a AlluxioFileUtils) ExecMountScripts() error {
	// Note: this script is mounted in master/statefulset.yaml
	command := []string{"/etc/fluid/scripts/mount.sh"}
	stdout, stderr, err := a.exec(command, true)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return err
	}
	return nil
}
