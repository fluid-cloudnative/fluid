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

func RemoveSymlink(targetPath string) (bool, error) {
	f, err := os.Lstat(targetPath)
	if err != nil {
		return false, fmt.Errorf("lstat targetPath %s error %v", targetPath, err)
	}
	// remove if targetPath is a symlink
	if f.Mode()&os.ModeSymlink != 0 {
		glog.Infof("%v is a symlink", targetPath)
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to remove symlink targetPath %s, error %v", targetPath, err)
		}
		// return true if and only if remove symlink successfully
		return true, nil
	}
	return false, nil
}
