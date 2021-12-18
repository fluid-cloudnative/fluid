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

package mountinfo

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/golang/glog"
	"path"
	"strings"
)

type MountPoint struct {
	SourcePath     string
	MountPath      string
	FilesystemType string
	ReadOnly       bool
}

func GetBrokenMountPoints() ([]MountPoint, error) {
	// get mountinfo from proc
	mountByPath, err := loadMountInfo()
	if err != nil {
		return nil, err
	}

	// get global mount set in map
	globalMountByName, err := getGlobalMounts(mountByPath)
	if err != nil {
		return nil, err
	}

	// get bind mount
	bindMountByName := getBindMounts(mountByPath)

	// get broken bind mount
	return getBrokenBindMounts(globalMountByName, bindMountByName), nil
}

func getGlobalMounts(mountByPath map[string]*Mount) (globalMountByName map[string]*Mount, err error) {
	globalMountByName = make(map[string]*Mount)
	// get fluid MountRoot
	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		return nil, err
	}

	for k, v := range mountByPath {
		if strings.Contains(k, mountRoot) {
			fields := strings.Split(k, "/")
			if len(fields) < 6 {
				continue
			}
			// fluid global mount path is: /{rootPath}/{runtimeType}/{namespace}/{runtimeName}/{runtimeTypeFuse}
			namespace, runtimeName := fields[3], fields[4]
			namespacedName := fmt.Sprintf("%s-%s", namespace, runtimeName)
			globalMountByName[namespacedName] = v
		}
	}
	return
}

func getBindMounts(mountByPath map[string]*Mount) (bindMountByName map[string][]*Mount) {
	bindMountByName = make(map[string][]*Mount)
	for k, m := range mountByPath {
		var runtimeNamespacedName string
		if strings.Contains(k, "kubernetes.io~csi") && strings.Contains(k, "mount") {
			fields := strings.Split(k, "/")
			if len(fields) < 3 {
				continue
			}
			runtimeNamespacedName = fields[len(fields)-2]
			bindMountByName[runtimeNamespacedName] = append(bindMountByName[runtimeNamespacedName], m)
		}
	}
	return
}

func getBrokenBindMounts(globalMountByName map[string]*Mount, bindMountByName map[string][]*Mount) (brokenMounts []MountPoint) {
	for name, bindMounts := range bindMountByName {
		globalMount, ok := globalMountByName[name]
		if !ok {
			// globalMount is unmount, ignore
			glog.V(4).Infof("ignoring mountpoint %s because of not finding its global mount point", name)
			continue
		}
		for _, bindMount := range bindMounts {
			if *bindMount.PeerGroup != *globalMount.PeerGroup {
				brokenMounts = append(brokenMounts, MountPoint{
					SourcePath:     path.Join(globalMount.MountPath, bindMount.Subtree),
					MountPath:      bindMount.MountPath,
					FilesystemType: bindMount.FilesystemType,
					ReadOnly:       bindMount.ReadOnly,
				})
			}
		}
	}
	return
}
