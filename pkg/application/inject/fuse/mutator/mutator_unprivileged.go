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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var fuseDeviceResourceName string

var (
	// datavolume-, volume-localtime for JindoFS
	// mem, ssd, hdd for Alluxio and GooseFS
	// cache-dir for JuiceFS
	cacheDirNames = []string{"datavolume-", "volume-localtime", "cache-dir", "mem", "ssd", "hdd"}

	// hostpath fuse mount point for Alluxio, JindoFS, GooseFS and JuiceFS
	hostMountNames = []string{"alluxio-fuse-mount", "jindofs-fuse-mount", "goosefs-fuse-mount", "juicefs-fuse-mount", "thin-fuse-mount", "efc-fuse-mount", "efc-sock"}

	// fuse devices for Alluxio, JindoFS, GooseFS
	hostFuseDeviceNames = []string{"alluxio-fuse-device", "jindofs-fuse-device", "goosefs-fuse-device", "thin-fuse-device"}
)

func init() {
	fuseDeviceResourceName = utils.GetStringValueFromEnv(common.EnvFuseDeviceResourceName, common.DefaultFuseDeviceResourceName)
}

type UnprivilegedMutator struct {
	DefaultMutator
}

var _ FluidObjectMutator = &UnprivilegedMutator{}

func NewUnprivilegedMutator(ctx MutatingContext, client client.Client, log logr.Logger) FluidObjectMutator {
	return &UnprivilegedMutator{
		DefaultMutator: DefaultMutator{
			MutatingContext: ctx,
			client:          client,
			log:             log,
		},
	}
}

func (mutator *UnprivilegedMutator) Mutate() (*MutatingPodSpecs, error) {
	if err := mutator.prepareResources(); err != nil {
		return nil, err
	}

	if err := mutator.mutateDatasetVolume(); err != nil {
		return nil, err
	}

	// 3. append volumes required by fuse sidecar container
	if err := mutator.appendFuseSidecarVolumes(); err != nil {
		return nil, err
	}

	if mutator.DatasetUsedInContainers {
		if err := mutator.prependFuseContainer(false /* asInit */); err != nil {
			return nil, err
		}
	}

	if mutator.DatasetUsedInInitContainers {
		if err := mutator.prependFuseContainer(true /* asInit */); err != nil {
			return nil, err
		}
	}

	return mutator.Specs, nil
}

func (mutator *UnprivilegedMutator) prepareResources() error {
	if err := mutator.prepareFuseTemplateForUnprivilegedPlatform(); err != nil {
		return err
	}

	if err := mutator.prepareUnprivilegedFuseContainerPostStartScript(); err != nil {
		return err
	}

	return nil
}

func (mutator *UnprivilegedMutator) prepareFuseTemplateForUnprivilegedPlatform() error {
	template := mutator.Template
	// remove the fuse related volumes if using virtual fuse device
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostMountNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostMountNames)

	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostFuseDeviceNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostFuseDeviceNames)

	// add virtual fuse device resource
	if template.FuseContainer.Resources.Limits == nil {
		template.FuseContainer.Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Limits[corev1.ResourceName(fuseDeviceResourceName)] = resource.MustParse("1")

	if template.FuseContainer.Resources.Requests == nil {
		template.FuseContainer.Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Requests[corev1.ResourceName(fuseDeviceResourceName)] = resource.MustParse("1")

	// invalidate privileged fuse container
	if template.FuseContainer.SecurityContext != nil {
		privilegedContainer := false
		template.FuseContainer.SecurityContext.Privileged = &privilegedContainer
		if template.FuseContainer.SecurityContext.Capabilities != nil {
			template.FuseContainer.SecurityContext.Capabilities.Add = utils.TrimCapabilities(template.FuseContainer.SecurityContext.Capabilities.Add, []string{"SYS_ADMIN"})
		}
	}

	return nil
}

func (mutator *DefaultMutator) prepareUnprivilegedFuseContainerPostStartScript() error {
	var (
		datasetName      = mutator.RuntimeInfo.GetName()
		datasetNamespace = mutator.RuntimeInfo.GetNamespace()
		fuseMountInfo    = mutator.Template.FuseMountInfo
	)

	dataset, err := utils.GetDataset(mutator.client, datasetName, datasetNamespace)
	if err != nil {
		return errors.Wrapf(err, "fail to get dataset \"%s/%s\" when preparing post start script for pvc %s", mutator.RuntimeInfo.GetNamespace(), mutator.RuntimeInfo.GetName(), mutator.PvcName)
	}

	// ownerReference := transfromer.GenerateOwnerReferenceFromObject(dataset)

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	mountPathInContainer := ""
	// mountPathInContainer := volumeMountInContainer.MountPath
	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      datasetName,
		Namespace: datasetNamespace,
	}, mountPathInContainer, fuseMountInfo.FsType, fuseMountInfo.SubPath, mutator.Options)

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

	mutator.Template.FuseContainer.VolumeMounts = append(mutator.Template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if mutator.Template.FuseContainer.Lifecycle == nil {
		mutator.Template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	mutator.Template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()
	mutator.Template.VolumesToAdd = append(mutator.Template.VolumesToAdd, gen.GetVolume())

	return nil
}
