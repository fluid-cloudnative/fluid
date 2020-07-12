package csi

import (
	"errors"
	"os/exec"

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

func checkMountReady(fluidPath string) error {
	glog.Infof("Try to check if the mount target %s is ready", fluidPath)
	if fluidPath == "" {
		return errors.New("target is not specified for checking the mount")
	}

	args := []string{fluidPath}
	command := exec.Command("/usr/local/bin/check_mount.sh", args...)
	glog.Infoln(command)
	stdoutStderr, err := command.CombinedOutput()
	glog.Infoln(string(stdoutStderr))
	return err
}
