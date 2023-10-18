package mutator

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type UnprivilegedMutator struct {
	// UnprivilegedMutator inherits from DefaultMutator
	DefaultMutator
}

var _ Mutator = &UnprivilegedMutator{}

func NewUnprivilegedMutator(opts MutatorBuildOpts) Mutator {
	return &UnprivilegedMutator{
		DefaultMutator: DefaultMutator{
			options: opts.Options,
			client:  opts.Client,
			log:     opts.Log,
			Specs:   opts.Specs,
		},
	}
}

func (mutator *UnprivilegedMutator) MutateWithRuntimeInfo(pvcName string, runtimeInfo base.RuntimeInfoInterface, nameSuffix string) error {
	template, err := runtimeInfo.GetFuseContainerTemplate()
	if err != nil {
		return errors.Wrapf(err, "failed to get fuse container template for runtime \"%s/%s\"", runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	helper := unprivilegedMutatorHelper{
		defaultMutatorHelper: defaultMutatorHelper{
			pvcName:     pvcName,
			template:    template,
			options:     mutator.options,
			runtimeInfo: runtimeInfo,
			nameSuffix:  nameSuffix,
			client:      mutator.client,
			log:         mutator.log,
			Specs:       mutator.Specs,
			ctx:         mutatingContext{},
		},
	}

	if err := helper.PrepareMutation(); err != nil {
		return errors.Wrapf(err, "failed to prepare mutation for runtime \"%s/%s\"", runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	_, err = helper.Mutate()
	if err != nil {
		return errors.Wrapf(err, "failed to mutate for runtime \"%s/%s\"", runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	return nil
}

func (mutator *UnprivilegedMutator) PostMutate() error {
	return mutator.DefaultMutator.PostMutate()
}

func (mutator *UnprivilegedMutator) GetMutatedPodSpecs() *MutatingPodSpecs {
	return mutator.DefaultMutator.GetMutatedPodSpecs()
}

type unprivilegedMutatorHelper struct {
	defaultMutatorHelper
}

func (mutator *unprivilegedMutatorHelper) PrepareMutation() error {
	if !mutator.options.EnableCacheDir {
		mutator.transformTemplateWithCacheDirDisabled()
	}

	mutator.transformTemplateWithUnprivilegedSidecarEnabled()

	if !mutator.options.SkipSidecarPostStartInject {
		if err := mutator.prepareFuseContainerPostStartScript(); err != nil {
			return err
		}
	}

	return nil
}

func (mutator *unprivilegedMutatorHelper) Mutate() (*MutatingPodSpecs, error) {
	return mutator.defaultMutatorHelper.Mutate()
}

func (mutator *unprivilegedMutatorHelper) prepareFuseContainerPostStartScript() error {
	// 4. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	// Post start script varies according to privileged or unprivileged sidecar.
	var (
		info             = mutator.runtimeInfo
		template         = mutator.template
		datasetName      = info.GetName()
		datasetNamespace = info.GetNamespace()
	)

	dataset, err := utils.GetDataset(mutator.client, datasetName, datasetNamespace)
	if err != nil {
		return err
	}

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	// Fluid assumes pvc name is the same with runtime's name
	gen := poststart.NewUnprivilegedPostStartScriptGenerator()
	cmKey := gen.GetConfigMapKeyByOwner(types.NamespacedName{Namespace: datasetNamespace, Name: datasetName}, template.FuseMountInfo.FsType)
	cm := gen.BuildConfigMap(ownerReference, cmKey)

	found, err := kubeclient.IsConfigMapExist(mutator.client, cmKey.Name, cmKey.Namespace)
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

	template.FuseContainer.VolumeMounts = append(template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if template.FuseContainer.Lifecycle == nil {
		template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()
	template.VolumesToAdd = append(template.VolumesToAdd, gen.GetVolume(cmKey))

	return nil
}
