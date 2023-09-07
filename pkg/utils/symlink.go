package utils

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateSymlink(targetPath, mountPath string) error {
	_, err := os.Lstat(targetPath)
	// If the target path does not exist, it will be created when create symlink
	if err != nil {
		if !os.IsNotExist(err) {
			glog.Errorf("Failed to lstat targetPath %s error %v", targetPath, err)
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		// symlink would create targetPath so delete it first
		glog.Infof("Deleting the targetPath before create symlink %v", targetPath)
		err := os.Remove(targetPath)
		if err != nil && !os.IsNotExist(err) {
			glog.Errorf("Failed to delete the target path %s error %v", targetPath, err)
			return status.Error(codes.Internal, fmt.Sprintf("Failed to delete the target path %s before create symlink, error %v", targetPath, err))
		}
	}
	// create symlink
	symlinkErr := os.Symlink(mountPath, targetPath)
	if symlinkErr != nil {
		glog.Errorf("Failed to create symlink %s link to %s, error %v", targetPath, mountPath, symlinkErr)
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create symlink %s -> %s, error %v", targetPath, mountPath, symlinkErr))
	}
	glog.Infof("Creating symlink %s link to %s successfully", targetPath, mountPath)
	return nil
}
