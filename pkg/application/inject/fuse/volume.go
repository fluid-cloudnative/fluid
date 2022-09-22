package fuse

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// overrideDatasetVolumes overrides any PersistentVolumeClaim volume that possesses a claimName equals to datasetPvcName with newDatasetVolume.
// it returns the affected volumes' name and the volumes after overriding.
func overrideDatasetVolumes(volumes []corev1.Volume, datasetPvcName string, newDatasetVolume corev1.Volume) ([]string, []corev1.Volume) {
	var datasetVolumeNames []string

	for i, volume := range volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == datasetPvcName {
			name := volume.Name
			volumes[i] = newDatasetVolume
			volumes[i].Name = name
			datasetVolumeNames = append(datasetVolumeNames, name)
		}
	}

	return datasetVolumeNames, volumes
}

// appendVolumes adds new volumes from volumesToAdd into existing volumes. It also resolve volume name conflicts when appending volumes.
// The func returns conflict names with mappings from old name to new name and the appended volumes.
func appendVolumes(volumes []corev1.Volume, volumesToAdd []corev1.Volume, namespacedName types.NamespacedName) (volumeNamesConflict map[string]string, retVolumes []corev1.Volume, err error) {
	// collect all volumes' names
	var volumeNames []string
	for _, volume := range volumes {
		volumeNames = append(volumeNames, volume.Name)
	}

	volumeNamesConflict = map[string]string{}
	// Append volumes
	if len(volumesToAdd) > 0 {
		log.V(1).Info("Before append volume", "original", volumes)
		// volumes = append(volumes, template.VolumesToAdd...)
		for _, volumeToAdd := range volumesToAdd {
			if !utils.ContainsString(volumeNames, volumeToAdd.Name) {
				volumes = append(volumes, volumeToAdd)
			} else {
				// Found conflict volume name
				newVolumeName, err := randomizeNewVolumeName(volumeToAdd.Name, volumeNames, namespacedName)
				if err != nil {
					return volumeNamesConflict, volumes, err
				}

				volume := volumeToAdd.DeepCopy()
				volume.Name = newVolumeName
				volumeNamesConflict[volumeToAdd.Name] = volume.Name
				volumeNames = append(volumeNames, newVolumeName)
				volumes = append(volumes, *volume)
			}
		}

		log.V(1).Info("After append volume", "original", volumes)
	}

	return volumeNamesConflict, volumes, nil
}

func randomizeNewVolumeName(origName string, existingNames []string, namespacedName types.NamespacedName) (string, error) {
	i := 0
	newVolumeName := utils.ReplacePrefix(origName, common.Fluid)
	for {
		if !utils.ContainsString(existingNames, newVolumeName) {
			break
		} else {
			if i > 100 {
				return "", fmt.Errorf("retry  the volume name %v for object %v because duplicate name more than 100 times, then give up", newVolumeName, types.NamespacedName{
					Namespace: namespacedName.Namespace,
					Name:      namespacedName.Name,
				})
			}
			suffix := common.Fluid + "-" + utils.RandomAlphaNumberString(3)
			newVolumeName = utils.ReplacePrefix(origName, suffix)
			log.Info("retry  the volume name because duplicate name",
				"name", namespacedName.Name,
				"namespace", namespacedName.Namespace,
				"volumeName", newVolumeName)
			i++
		}
	}

	return newVolumeName, nil
}
