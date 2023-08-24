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

package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/validation"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/mount"
)

const MountRoot string = "MOUNT_ROOT"

// GetMountRoot gets the value of the env variable named MOUNT_ROOT
func GetMountRoot() (string, error) {
	// mountRoot := os.Getenv(MountRoot)
	mountRoot := "/runtime-mnt"

	if err := validation.IsValidMountRoot(mountRoot); err != nil {
		return mountRoot, err
	}
	return mountRoot, nil
}

func CheckMountReadyAndSubPathExist(fluidPath string, mountType string, subPath string) (err error) {
	glog.Infof("Try to check if the mount target %s is ready", fluidPath)
	if fluidPath == "" {
		return errors.New("target is not specified for checking the mount")
	}
	args := []string{fluidPath, mountType, subPath}
	command, err := cmdguard.Command("/usr/local/bin/check_mount.sh", args...)
	if err != nil {
		return
	}
	glog.Infoln(command)
	stdoutStderr, err := command.CombinedOutput()
	glog.Infoln(string(stdoutStderr))

	if err != nil {
		var checkMountErr *exec.ExitError
		if errors.As(err, &checkMountErr) {
			switch checkMountErr.ExitCode() {
			case 1:
				// exitcode=1 indicates timeout waiting for mount point to be ready
				return errors.New("timeout waiting for FUSE mount point to be ready")
			case 2:
				// exitcode=2 indicates subPath not exists
				return fmt.Errorf("subPath \"%s\" not exists under FUSE mount", subPath)
			}
		}
		return err
	}
	return nil
}

func IsMounted(absPath string) (bool, error) {
	glog.Infof("abspath: %s\n", absPath)
	_, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}

	file, err := os.ReadFile("/proc/mounts")
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
