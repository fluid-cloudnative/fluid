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
	"path"
	"strings"

	"github.com/golang/glog"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

type MountPoint struct {
	SourcePath            string
	MountPath             string
	FilesystemType        string
	ReadOnly              bool
	Count                 int
	NamespacedDatasetName string // <namespace>-<dataset>
}

func GetBrokenMountPoints() (MountPoints, error) {
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
			// fluid global mount path is: /{rootPath}/{runtimeType}/{namespace}/{datasetName}/{runtimeTypeFuse}
			namespace, datasetName := fields[3], fields[4]
			namespacedName := fmt.Sprintf("%s-%s", namespace, datasetName)
			globalMountByName[namespacedName] = v
		}
	}
	return
}

func getBindMounts(mountByPath map[string]*Mount) (bindMountByName map[string][]*Mount) {
	bindMountByName = make(map[string][]*Mount)
	for k, m := range mountByPath {
		var datasetNamespacedName string
		mCopy := m
		if strings.Contains(k, "kubernetes.io~csi") && strings.Contains(k, "mount") {
			// fluid bind mount target path is: /{kubeletRootDir}(default: /var/lib/kubelet)/pods/{podUID}/volumes/kubernetes.io~csi/{namespace}-{datasetName}/mount
			fields := strings.Split(k, "/")
			if len(fields) < 3 {
				continue
			}
			datasetNamespacedName = fields[len(fields)-2]
			bindMountByName[datasetNamespacedName] = append(bindMountByName[datasetNamespacedName], mCopy)
		}
		if strings.Contains(k, "volume-subpaths") {
			// pod using subPath, bind mount path is: /{kubeletRootDir}(default: /var/lib/kubelet)/pods/{podUID}/volume-subpaths/{namespace}-{datasetName}/{containerName}/{volumeIndex}
			fields := strings.Split(k, "/")
			if len(fields) < 4 {
				continue
			}
			datasetNamespacedName = fields[len(fields)-3]
			bindMountByName[datasetNamespacedName] = append(bindMountByName[datasetNamespacedName], mCopy)
		}
	}
	return
}

func getBrokenBindMounts(globalMountByName map[string]*Mount, bindMountByName map[string][]*Mount) (brokenMounts []MountPoint) {
	for name, bindMounts := range bindMountByName {
		globalMount, ok := globalMountByName[name]
		if !ok {
			// globalMount is unmount, ignore
			glog.V(6).Infof("ignoring mountpoint %s because of not finding its global mount point", name)
			continue
		}
		for _, bm := range bindMounts {
			bindMount := bm

			if strings.HasSuffix(bindMount.Subtree, "//deleted") {
				glog.V(5).Infof("bindMount subtree has been deleted, trim /deleted suffix, bindmount: %v", bindMount)
				bindMount.Subtree = strings.TrimSuffix(bindMount.Subtree, "//deleted")
			}

			// In case of not sharing same peer group in mount info, meaning it a broken mount point
			if len(utils.IntersectIntegerSets(bindMount.PeerGroups, globalMount.PeerGroups)) == 0 {
				brokenMounts = append(brokenMounts, MountPoint{
					SourcePath:            path.Join(globalMount.MountPath, bindMount.Subtree),
					MountPath:             bindMount.MountPath,
					FilesystemType:        bindMount.FilesystemType,
					ReadOnly:              bindMount.ReadOnly,
					Count:                 bindMount.Count,
					NamespacedDatasetName: name,
				})
			}
		}
	}
	return
}

type MountPoints []MountPoint

func (mp MountPoints) Len() int { return len(mp) }
func (mp MountPoints) Less(i, j int) bool {
	if strings.Contains(mp[i].MountPath, "main") && !strings.Contains(mp[j].MountPath, "main") {
		return false
	}
	if !strings.Contains(mp[i].MountPath, "main") && strings.Contains(mp[j].MountPath, "main") {
		return true
	}
	if strings.Contains(mp[i].MountPath, "subpath") && !strings.Contains(mp[j].MountPath, "subpath") {
		return false
	}
	if !strings.Contains(mp[i].MountPath, "subpath") && strings.Contains(mp[j].MountPath, "subpath") {
		return true
	}
	return mp[i].MountPath < mp[j].MountPath
}
func (mp MountPoints) Swap(i, j int) {
	mp[i].MountPath, mp[j].MountPath = mp[j].MountPath, mp[i].MountPath
}
