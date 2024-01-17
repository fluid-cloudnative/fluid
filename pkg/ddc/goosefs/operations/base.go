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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
)

type GooseFSFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewGooseFSFileUtils(podName string, containerName string, namespace string, log logr.Logger) GooseFSFileUtils {

	return GooseFSFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// IsExist checks if the goosefsPath exists
func (a GooseFSFileUtils) IsExist(goosefsPath string) (found bool, err error) {
	var (
		command = []string{"goosefs", "fs", "ls", goosefsPath}
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

// Get summary info of the GooseFS Engine
func (a GooseFSFileUtils) ReportSummary() (summary string, err error) {
	var (
		command = []string{"goosefs", "fsadmin", "report", "summary"}
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
func (a GooseFSFileUtils) LoadMetadataWithoutTimeout(goosefsPath string) (err error) {
	var (
		command = []string{"goosefs", "fs", "loadMetadata", "-R", goosefsPath}
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
func (a GooseFSFileUtils) LoadMetaData(goosefsPath string, sync bool) (err error) {
	var (
		// command = []string{"goosefs", "fs", "-Dgoosefs.user.file.metadata.sync.interval=0", "ls", "-R", goosefsPath}
		// command = []string{"goosefs", "fs", "-Dgoosefs.user.file.metadata.sync.interval=0", "count", goosefsPath}
		command []string
		stdout  string
		stderr  string
	)

	if sync {
		command = []string{"goosefs", "fs", "-Dgoosefs.user.file.metadata.sync.interval=0", "ls", "-R", goosefsPath}
	} else {
		command = []string{"goosefs", "fs", "ls", "-R", goosefsPath}
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

// QueryMetadataInfoFile query the metadata info file.
func (a GooseFSFileUtils) QueryMetaDataInfoIntoFile(key KeyOfMetaDataFile, filename string) (value string, err error) {
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

func (a GooseFSFileUtils) Mkdir(goosefsPath string) (err error) {
	var (
		command = []string{"goosefs", "fs", "mkdir", goosefsPath}
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

func (a GooseFSFileUtils) Mount(goosefsPath string,
	ufsPath string,
	options map[string]string,
	readOnly bool,
	shared bool) (err error) {

	// exist, expectedErr := a.IsExist(goosefsPath)
	// if expectedErr != nil {
	// 	return expectedErr
	// }

	// if !exist {
	// 	expectedErr = a.Mkdir(goosefsPath)
	// 	if expectedErr != nil {
	// 		return expectedErr
	// 	}
	// }

	var (
		command = []string{"goosefs", "fs", "mount"}
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

	command = append(command, goosefsPath, ufsPath)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}

// UnMount execute command `goosefs fs umount $path` to unmount mountpoint
func (a GooseFSFileUtils) UnMount(goosefsPath string) (err error) {
	var (
		command = []string{"goosefs", "fs", "unmount"}
		stderr  string
		stdout  string
	)

	command = append(command, goosefsPath)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}

func (a GooseFSFileUtils) IsMounted(goosefsPath string) (mounted bool, err error) {
	var (
		command = []string{"goosefs", "fs", "mount"}
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
		a.log.Info("parse output of isMounted", "goosefsPath", goosefsPath, "fields", fields)
		if fields[2] == goosefsPath {
			mounted = true
			return mounted, nil
		}
	}

	// pattern := fmt.Sprintf(" on %s ", goosefsPath)
	// if strings.Contains(stdout, pattern) {
	// 	mounted = true
	// }

	return mounted, err
}

// Check if the GooseFS is ready by running `goosefs fsadmin report` command
func (a GooseFSFileUtils) Ready() (ready bool) {
	var (
		command = []string{"goosefs", "fsadmin", "report"}
	)

	_, _, err := a.exec(command, true)
	if err == nil {
		ready = true
	}

	return ready
}

func (a GooseFSFileUtils) Du(goosefsPath string) (ufs int64, cached int64, cachedPercentage string, err error) {
	var (
		command = []string{"goosefs", "fs", "du", "-s", goosefsPath}
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

// The count of the GooseFS Filesystem
func (a GooseFSFileUtils) Count(goosefsPath string) (fileCount int64, folderCount int64, total int64, err error) {
	var (
		command                          = []string{"goosefs", "fs", "count", goosefsPath}
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

// file count of the GooseFS Filesystem (except folder)
// use "goosefs fsadmin report metrics" for better performance
func (a GooseFSFileUtils) GetFileCount() (fileCount int64, err error) {
	args := []string{"goosefs", "fsadmin", "report", "metrics", "|", "grep", "Master.FilesCompleted"}
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

// ReportMetrics get goosefs metrics by running `goosefs fsadmin report metrics` command
func (a GooseFSFileUtils) ReportMetrics() (metrics string, err error) {
	var (
		command = []string{"goosefs", "fsadmin", "report", "metrics"}
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

// ReportCapacity get goosefs capacity info by running `goosefs fsadmin report capacity` command
func (a GooseFSFileUtils) ReportCapacity() (report string, err error) {
	var (
		command = []string{"goosefs", "fsadmin", "report", "capacity"}
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

func (a GooseFSFileUtils) MasterPodName() (masterPodName string, err error) {
	var (
		command = []string{"goosefs", "fsadmin", "report"}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = a.exec(command, true)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return stdout, err
	}

	str := strings.Split(stdout, "\n")
	data := strings.Fields(str[1])
	address := strings.Split(data[2], ":")[0]

	return address, nil
}

// exec with timeout
func (a GooseFSFileUtils) exec(command []string, verbose bool) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = a.execWithoutTimeout(command, verbose)
		ch <- "done"
	}()

	select {
	case <-ch:
		a.log.Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}

	return
}

// execWithoutTimeout
func (a GooseFSFileUtils) execWithoutTimeout(command []string, verbose bool) (stdout string, stderr string, err error) {
	err = utils.ValidateCommandSlice(command)
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
