/*
Copyright 2021 The Fluid Authors.

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
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/security"
)

type JuiceFileUtils struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewJuiceFileUtils(podName string, containerName string, namespace string, log logr.Logger) JuiceFileUtils {
	return JuiceFileUtils{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// Load the metadata without timeout
func (j JuiceFileUtils) LoadMetadataWithoutTimeout(juicefsPath string) (err error) {
	var (
		command = []string{"ls", "-lR", juicefsPath}
		stdout  string
		stderr  string
	)

	start := time.Now()
	stdout, stderr, err = j.execWithoutTimeout(command)
	duration := time.Since(start)
	j.log.Info("Async Load Metadata took times to run", "period", duration)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	} else {
		j.log.Info("Async Load Metadata finished", "stdout", stdout)
	}
	return
}

// The count of the JuiceFS Filesystem
func (j JuiceFileUtils) Count(juiceSubPath string) (total int64, err error) {
	var (
		command = []string{"du", "-sb", juiceSubPath}
		stdout  string
		stderr  string
		utotal  int64
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	// [File Count Folder Count Total Bytes 1152 4 154262709011]
	str := strings.Split(stdout, "\n")

	if len(str) != 1 {
		err = fmt.Errorf("failed to parse %s in Count method", str)
		return
	}

	data := strings.Fields(str[0])
	if len(data) != 2 {
		err = fmt.Errorf("failed to parse %s in Count method", data)
		return
	}

	utotal, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return
	}
	if utotal < 0 {
		err = fmt.Errorf("the return value of Count method is negative")
		return
	}

	return utotal, err
}

// file count of the JuiceFS Filesystem (except folder)
// use "ls -lR  xxx|grep "^-"| wc -l"
func (j JuiceFileUtils) GetFileCount(juiceSubPath string) (fileCount int64, err error) {
	var (
		//strs    = "du -ah juiceSubPath |grep ^- |wc -l "
		strs    = fmt.Sprintf("ls -lR %s |grep ^- |wc -l ", security.EscapeBashStr(juiceSubPath))
		command = []string{"bash", "-c", strs}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = j.exec(command)
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

// Mkdir mkdir in juicefs container
func (j JuiceFileUtils) Mkdir(juiceSubPath string) (err error) {
	var (
		command = []string{"mkdir", juiceSubPath}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		if strings.Contains(stdout, "File exists") {
			err = nil
		} else {
			err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
			return
		}
	}
	return
}

// DeleteCacheDirs delete cache dir in pod
func (j JuiceFileUtils) DeleteCacheDirs(dirs []string) (err error) {
	for _, dir := range dirs {
		// cache dir check
		match := ValidCacheDir(dir)
		if !match {
			j.log.Info("invalid cache directory, skip cleaning up", "cacheDir", dir)
			return
		}
	}
	var (
		command = []string{"rm", "-rf"}
		stdout  string
		stderr  string
	)
	command = append(command, dirs...)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	return
}

// DeleteCacheDir delete cache dir in pod
func (j JuiceFileUtils) DeleteCacheDir(dir string) (err error) {
	// cache dir check
	match := ValidCacheDir(dir)
	if !match {
		j.log.Info("invalid cache directory, skip cleaning up", "cacheDir", dir)
		return
	}
	var (
		command = []string{"rm", "-rf", dir}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	return
}

// GetStatus get status of volume
func (j JuiceFileUtils) GetStatus(source string) (status string, err error) {
	var (
		command = []string{"sh", "-c", fmt.Sprintf("juicefs status %s", source)}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	status = stdout
	return
}

// GetMetric Get pod metrics
func (j JuiceFileUtils) GetMetric(juicefsPath string) (metrics string, err error) {
	var (
		command = []string{"cat", fmt.Sprintf("%s/%s", juicefsPath, ".stats")}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	metrics = stdout
	return
}

// GetUsedSpace Get used space in byte
// equal to `df --block-size=1 | grep juicefsPath`
func (j JuiceFileUtils) GetUsedSpace(juicefsPath string) (usedSpace int64, err error) {
	var (
		command = []string{"df", "--block-size=1"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	var str string
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, juicefsPath) {
			str = line
			break
		}
	}
	// [<Filesystem>       <Size>  <Used> <Avail> <Use>% <Mounted on>]
	data := strings.Fields(str)
	if len(data) != 6 {
		err = fmt.Errorf("failed to parse %s in GetUsedSpace method", data)
		return
	}

	usedSpace, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return
	}

	return usedSpace, err
}

// exec with timeout
func (j JuiceFileUtils) exec(command []string) (stdout string, stderr string, err error) {
	j.log.Info("execute begin", "command", command)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1500)
	ch := make(chan string, 1)
	defer cancel()

	go func() {
		stdout, stderr, err = j.execWithoutTimeout(command)
		ch <- "done"
	}()

	select {
	case <-ch:
		j.log.Info("execute in time", "command", command)
	case <-ctx.Done():
		err = fmt.Errorf("timeout when executing %v", command)
	}
	return
}

// execWithoutTimeout
func (j JuiceFileUtils) execWithoutTimeout(command []string) (stdout string, stderr string, err error) {
	// validate the pipe command with white list
	err = cmdguard.ValidateCommandSlice(command)
	if err != nil {
		return
	}

	stdout, stderr, err = kubeclient.ExecCommandInContainer(j.podName, j.container, j.namespace, command)
	if err != nil {
		j.log.Info("Stdout", "Command", command, "Stdout", stdout)
		j.log.Error(err, "Failed", "Command", command, "FailedReason", stderr)
		return
	}
	j.log.V(1).Info("Stdout", "Command", command, "Stdout", stdout)
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
func (j JuiceFileUtils) QueryMetaDataInfoIntoFile(key KeyOfMetaDataFile, filename string) (value string, err error) {
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
		j.log.Error(errors.New("the key not in  metadatafile"), "key", key)
	}
	var (
		str     = "'" + line + "' " + filename
		command = []string{"sed", "-n", str}
		stdout  string
		stderr  string
	)
	stdout, stderr, err = j.exec(command)
	if err != nil {
		err = fmt.Errorf("execute command %v with  expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
	} else {
		value = strings.TrimPrefix(stdout, string(key)+": ")
	}
	return
}

func ValidCacheDir(dir string) (match bool) {
	return strings.HasSuffix(dir, "raw/chunks")
}
