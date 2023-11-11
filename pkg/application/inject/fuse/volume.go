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

package fuse

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// appendVolumes adds new volumes from volumesToAdd into existing volumes. It also resolve volume name conflicts when appending volumes.
// The func returns conflict names with mappings from old name to new name and the appended volumes.
func (s *Injector) appendVolumes(volumes []corev1.Volume, volumesToAdd []corev1.Volume, nameSuffix string) (volumeNamesConflict map[string]string, retVolumes []corev1.Volume, err error) {
	// collect all volumes' names
	var volumeNames []string
	for _, volume := range volumes {
		volumeNames = append(volumeNames, volume.Name)
	}

	volumeNamesConflict = map[string]string{}
	// Append volumes
	if len(volumesToAdd) > 0 {
		s.log.V(1).Info("Before append volume", "original", volumes)
		// volumes = append(volumes, template.VolumesToAdd...)
		for _, volumeToAdd := range volumesToAdd {
			// nameSuffix would be like: "-0", "-1", "-2", "-3", ...
			oldVolumeName := volumeToAdd.Name
			newVolumeName := volumeToAdd.Name + nameSuffix
			if utils.ContainsString(volumeNames, newVolumeName) {
				newVolumeName, err = s.randomizeNewVolumeName(newVolumeName, volumeNames)
				if err != nil {
					return volumeNamesConflict, volumes, err
				}
			}
			volumeToAdd.Name = newVolumeName
			volumeNames = append(volumeNames, newVolumeName)
			volumes = append(volumes, volumeToAdd)
			if oldVolumeName != newVolumeName {
				volumeNamesConflict[oldVolumeName] = newVolumeName
			}
		}

		s.log.V(1).Info("After append volume", "original", volumes)
	}

	return volumeNamesConflict, volumes, nil
}

func (s *Injector) randomizeNewVolumeName(origName string, existingNames []string) (string, error) {
	i := 0
	newVolumeName := utils.ReplacePrefix(origName, common.Fluid)
	for {
		if !utils.ContainsString(existingNames, newVolumeName) {
			break
		} else {
			if i > 100 {
				return "", fmt.Errorf("retry  the volume name %v because duplicate name more than 100 times, then give up", newVolumeName)
			}
			suffix := common.Fluid + "-" + utils.RandomAlphaNumberString(3)
			newVolumeName = utils.ReplacePrefix(origName, suffix)
			s.log.Info("retry  the volume name because duplicate name",
				"volumeName", newVolumeName)
			i++
		}
	}

	return newVolumeName, nil
}
