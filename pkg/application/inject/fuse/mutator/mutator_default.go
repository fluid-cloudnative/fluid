package mutator

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefaultMutator struct {
	MutatingContext
	client client.Client
	log    logr.Logger
}

var _ FluidObjectMutator = &DefaultMutator{}

func NewDefaultMutator(ctx MutatingContext, client client.Client, log logr.Logger) FluidObjectMutator {
	return &DefaultMutator{
		MutatingContext: ctx,
		client:          client,
		log:             log,
	}
}

func (mutator *DefaultMutator) Mutate() (*MutatingPodSpecs, error) {
	// 1. prepare platform-specific resources
	if err := mutator.prepareResources(); err != nil {
		return nil, err
	}

	// 2. mutate & append volumes
	if err := mutator.mutateDatasetVolume(); err != nil {
		return nil, err
	}

	// 3. append volumes required by fuse sidecar container
	if err := mutator.appendFuseSidecarVolumes(); err != nil {
		return nil, err
	}

	if mutator.datasetUsedInContainers {
		if err := mutator.prependFuseContainer(false /* asInit */); err != nil {
			return nil, err
		}
	}

	if mutator.datasetUsedInInitContainers {
		if err := mutator.prependFuseContainer(true /* asInit */); err != nil {
			return nil, err
		}
	}

	return mutator.specs, nil
}

func (mutator *DefaultMutator) prepareResources() error {
	return mutator.prepareFuseContainerPostStartScript()
}

func (mutator *DefaultMutator) prepareFuseContainerPostStartScript() error {
	var (
		datasetName      = mutator.runtimeInfo.GetName()
		datasetNamespace = mutator.runtimeInfo.GetNamespace()
		fuseMountInfo    = mutator.template.FuseMountInfo
	)

	dataset, err := utils.GetDataset(mutator.client, datasetName, datasetNamespace)
	if err != nil {
		return errors.Wrapf(err, "fail to get dataset \"%s/%s\" when preparing post start script for pvc %s", mutator.runtimeInfo.GetNamespace(), mutator.runtimeInfo.GetName(), mutator.pvcName)
	}

	// ownerReference := transfromer.GenerateOwnerReferenceFromObject(dataset)

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	volumeMountInContainer, err := kubeclient.GetFuseMountInContainer(mutator.template.FuseMountInfo.FsType, mutator.template.FuseContainer)
	if err != nil {
		return err
	}
	mountPathInContainer := volumeMountInContainer.MountPath
	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      datasetName,
		Namespace: datasetNamespace,
	}, mountPathInContainer, fuseMountInfo.FsType, fuseMountInfo.SubPath, mutator.options)

	cm := gen.BuildConfigmap(ownerReference)
	found, err := kubeclient.IsConfigMapExist(mutator.client, cm.Name, cm.Namespace)
	if err != nil {
		return err
	}

	if !found {
		err = mutator.client.Create(context.TODO(), cm)
		if err != nil {
			// If ConfigMap creation succeeds concurrently, continue to mutate
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				return err
			}
		}
	}

	mutator.template.FuseContainer.VolumeMounts = append(mutator.template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if mutator.template.FuseContainer.Lifecycle == nil {
		mutator.template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	mutator.template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()
	mutator.template.VolumesToAdd = append(mutator.template.VolumesToAdd, gen.GetVolume())

	return nil
}

func (mutator *DefaultMutator) mutateDatasetVolume() error {
	volumes := mutator.specs.Volumes

	mountPath := mutator.template.FuseMountInfo.MountPath
	if mutator.template.FuseMountInfo.SubPath != "" {
		mountPath = mountPath + "/" + mutator.template.FuseMountInfo.SubPath
	}

	mutatedDatasetVolume := corev1.Volume{
		Name: "",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: mountPath,
			},
		},
	}

	var overriddenVolumeNames []string
	for i, volume := range volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == mutator.pvcName {
			name := volume.Name
			volumes[i] = mutatedDatasetVolume
			volumes[i].Name = name
			overriddenVolumeNames = append(overriddenVolumeNames, name)
		}
	}

	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	for _, container := range mutator.specs.Containers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
			}
		}
	}

	for _, container := range mutator.specs.InitContainers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
			}
		}
	}

	return nil
}

func (mutator *DefaultMutator) appendFuseSidecarVolumes() error {
	// collect all volumes' names
	var (
		// volumeNames  = []string{}
		volumesToAdd = mutator.template.VolumesToAdd
		nameSuffix   = mutator.nameSuffix
	)
	// for _, volume := range mutator.specs.Volumes {
	// 	volumeNames = append(volumeNames, volume.Name)
	// }

	// Append volumes
	if len(volumesToAdd) > 0 {
		mutator.log.V(1).Info("Before append volume", "original", mutator.specs.Volumes)
		// volumes = append(volumes, template.VolumesToAdd...)
		for _, volumeToAdd := range volumesToAdd {
			// nameSuffix would be like: "-0", "-1", "-2", "-3", ...
			oldVolumeName := volumeToAdd.Name
			newVolumeName := volumeToAdd.Name + nameSuffix
			// if utils.ContainsString(volumeNames, newVolumeName) {
			// 	newVolumeName, err = s.randomizeNewVolumeName(newVolumeName, volumeNames)
			// 	if err != nil {
			// 		return volumeNamesConflict, volumes, err
			// 	}
			// }
			volumeToAdd.Name = newVolumeName
			// volumeNames = append(volumeNames, newVolumeName)
			mutator.specs.Volumes = append(mutator.specs.Volumes, volumeToAdd)
			if oldVolumeName != newVolumeName {
				mutator.appendedVolumeNames[oldVolumeName] = newVolumeName
			}
		}

		mutator.log.V(1).Info("After append volume", "original", mutator.specs.Volumes)
	}

	return nil
}

func (mutator *DefaultMutator) prependFuseContainer(asInit bool) error {
	fuseContainer := mutator.template.FuseContainer
	if !asInit {
		fuseContainer.Name = common.FuseContainerName + mutator.nameSuffix
	} else {
		fuseContainer.Name = common.InitFuseContainerName + mutator.nameSuffix
	}

	if asInit {
		fuseContainer.Lifecycle = nil
		fuseContainer.Command = []string{"sleep"}
		fuseContainer.Args = []string{"2s"}
	}

	for oldName, newName := range mutator.appendedVolumeNames {
		for i, volumeMount := range fuseContainer.VolumeMounts {
			if volumeMount.Name == oldName {
				fuseContainer.VolumeMounts[i].Name = newName
			}
		}
	}

	if !asInit {
		mutator.specs.Containers = append([]corev1.Container{fuseContainer}, mutator.specs.Containers...)
	} else {
		mutator.specs.InitContainers = append([]corev1.Container{fuseContainer}, mutator.specs.InitContainers...)
	}
	return nil
}
