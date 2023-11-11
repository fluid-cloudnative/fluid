/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/mount"
)

const MountRoot string = "MOUNT_ROOT"

// GetMountRoot gets the value of the env variable named MOUNT_ROOT
func GetMountRoot() (string, error) {
	mountRoot := os.Getenv(MountRoot)

	if !filepath.IsAbs(mountRoot) {
		return mountRoot, fmt.Errorf("the the value of the env variable named MOUNT_ROOT is illegal")
	}
	return mountRoot, nil
}

func CheckMountReadyAndSubPathExist(fluidPath string, mountType string, subPath string) error {
	glog.Infof("Try to check if the mount target %s is ready", fluidPath)
	if fluidPath == "" {
		return errors.New("target is not specified for checking the mount")
	}
	args := []string{fluidPath, mountType, subPath}
	command := exec.Command("/usr/local/bin/check_mount.sh", args...)
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
