package mutator

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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
	pvcName     string
	runtimeInfo base.RuntimeInfoInterface
	template    *common.FuseInjectionTemplate
	options     common.FuseSidecarInjectOption

	client client.Client
	log    logr.Logger
}

var _ FluidObjectMutator = &DefaultMutator{}

func (mutator *DefaultMutator) Mutate(pod common.FluidObject, template *common.FuseInjectionTemplate, options common.FuseSidecarInjectOption) error {
	// 1. prepare platform-specific resources
	if err := mutator.prepareResources(); err != nil {
		return err
	}

	// 2. mutate & append volumes

	// 3. mutate volume mounts
	return nil
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

func (mutator *DefaultMutator) mutateVolumes(pod common.FluidObject) error {
	volumes, err := pod.GetVolumes()
	if err != nil {
		return err
	}

	
}
