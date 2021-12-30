package csi

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

// func isMounted(target string) (bool, error) {
// 	if target == "" {
// 		return false, errors.New("target is not specified for checking the mount")
// 	}
// 	findmntCmd := "grep"
// 	findmntArgs := []string{target, "/proc/mounts"}
// 	out, err := exec.Command(findmntCmd, findmntArgs...).CombinedOutput()
// 	outStr := strings.TrimSpace(string(out))
// 	if err != nil {
// 		if outStr == "" {
// 			return false, nil
// 		}
// 		return false, fmt.Errorf("checking mounted failed: %v cmd: %q output: %q",
// 			err, findmntCmd, outStr)
// 	}
// 	if strings.Contains(outStr, target) {
// 		return true, nil
// 	}
// 	return false, nil
// }

func checkMountReady(fluidPath string, mountType string) error {
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

func isMounted(absPath string) (bool, error) {
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
