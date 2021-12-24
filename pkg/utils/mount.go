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

package utils

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const MountRoot string = "MOUNT_ROOT"

//GetMountRoot gets the value of the env variable named MOUNT_ROOT
func GetMountRoot() (string, error) {
	mountRoot := os.Getenv(MountRoot)

	if !filepath.IsAbs(mountRoot) {
		return mountRoot, fmt.Errorf("the the value of the env variable named MOUNT_ROOT is illegal")
	}
	return mountRoot, nil
}

func CheckMountReady(fluidPath string, mountType string) error {
	glog.Infof("Try to check if the mount target %s is ready", fluidPath)
	if fluidPath == "" {
		return errors.New("target is not specified for checking the mount")
	}
	args := []string{fluidPath, mountType}
	command := exec.Command("/usr/local/bin/check_mount.sh", args...)
	glog.Infoln(command)
	stdoutStderr, err := command.CombinedOutput()
	glog.Infoln(string(stdoutStderr))
	return err
}

func IsMounted(absPath string) (bool, error) {
	glog.Infof("abspath: %s\n", absPath)
	_, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}

	file, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		if len(tokens) < 2 {
			continue
		}
		if tokens[1] == absPath {
			glog.Infof("found mount info %s for %s", line, absPath)
			return true, nil
		}
	}
	return false, nil
}

func CheckMountPointBroken(mountPath string) (broken bool, err error) {
	if mountPath == "" {
		return false, errors.New("target is not specified for checking the mount")
	}
	existed, err := mount.PathExists(mountPath)
	if err != nil {
		if mount.IsCorruptedMnt(err) {
			return true, nil
		}
		return false, fmt.Errorf("checking mounted failed: %v", err)
	}
	if !existed {
		return false, fmt.Errorf("mountPath %s not exist", mountPath)
	}
	return false, nil
}

func GetRuntimeNameFromFusePod(pod corev1.Pod) (runtimeName string, err error) {
	podName := pod.Name
	strList := strings.Split(podName, "-fuse-")
	if len(strList) != 2 {
		err = fmt.Errorf("can't get runtime name from pod: %s, namespace: %s", pod.Name, pod.Namespace)
		return
	}
	runtimeName = strList[0]
	return
}

func IsFusePod(pod corev1.Pod) bool {
	labels := pod.Labels
	for k, v := range labels {
		if k == "role" && strings.Contains(v, "-fuse") {
			return true
		}
	}
	return false
}
