package dataset

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getDatasetRef(name, namespace string) string {
	// split by '#', can not use '-' because namespace or name can contain '-'
	return fmt.Sprintf("%s#%s", namespace, name)
}

func getPvName(name, namespace string) string {
	return fmt.Sprintf("%s-%s", name, namespace)
}

func getRuntimeType(dataset datav1alpha1.Dataset) (string, error) {
	if len(dataset.Status.Runtimes) != 0 {
		return dataset.Status.Runtimes[0].Type, nil
	}
	return "", fmt.Errorf("dataset %s do not bind the runtime", dataset.Name)
}

// TODO make configmap getter/setter to be a runtime interface
func createConfigMapForRefDataset(client client.Client, virtualDataset datav1alpha1.Dataset, runtime base.RuntimeInfoInterface) error {
	runtimeType := runtime.GetRuntimeType()

	physicalName := runtime.GetName()
	physicalNamespace := runtime.GetNamespace()
	virtualNameSpace := virtualDataset.GetNamespace()

	ownerReference := metav1.OwnerReference{
		APIVersion: virtualDataset.APIVersion,
		Kind:       virtualDataset.Kind,
		Name:       virtualDataset.Name,
		UID:        virtualDataset.UID,
	}

	// values configmap, each runtime will create it.
	valueConfigMapName := physicalName + "-" + runtimeType + "-values"
	// copy the configmap to virtual namespace with the same name.
	err := copyConfigMap(client, valueConfigMapName, physicalNamespace, virtualNameSpace, ownerReference)
	if err != nil {
		return err
	}

	// config configmap, each runtime except thin runtime will create it, but jindo runtime has different name.
	if runtimeType != common.ThinRuntime {
		configMapName := physicalName + "-config"
		if runtimeType == common.JindoRuntime {
			configMapName = physicalName + "-jindofs-config"
		}

		err = copyConfigMap(client, configMapName, physicalNamespace, virtualNameSpace, ownerReference)
		if err != nil {
			return err
		}
	}

	// jindo runtime extra configmaps
	if runtimeType == common.JindoRuntime {
		clientConfigMapName := physicalName + "-jindofs-client-config"
		err = copyConfigMap(client, clientConfigMapName, physicalNamespace, virtualNameSpace, ownerReference)
		if err != nil {
			return err
		}
	}

	// thin runtime extra configmaps
	if runtimeType == common.ThinRuntime {
		runtimesetConfigMapName := physicalName + "-runtimeset"
		err = copyConfigMap(client, runtimesetConfigMapName, physicalNamespace, virtualNameSpace, ownerReference)
		if err != nil {
			return err
		}
	}

	return err
}

func copyConfigMap(client client.Client, configMapName string, physicalNamespace string, virtualNameSpace string, reference metav1.OwnerReference) error {
	found, err := kubeclient.IsConfigMapExist(client, configMapName, virtualNameSpace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// copy configmap
	physicalValueCm, err := kubeclient.GetConfigmapByName(client, configMapName, physicalNamespace)
	if err != nil {
		return err
	}
	// if the physical dataset configmap not created, return error and requeue
	if physicalValueCm == nil {
		return fmt.Errorf("runtime configmap %s do not exist", configMapName)
	}
	// create the virtual dataset configmap if not exist
	copyCM := physicalValueCm.DeepCopy()

	virtualCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            copyCM.Name,
			Namespace:       virtualNameSpace,
			Labels:          copyCM.Labels,
			Annotations:     copyCM.Annotations,
			OwnerReferences: []metav1.OwnerReference{reference},
		},
		Data: copyCM.Data,
	}

	err = client.Create(context.TODO(), virtualCM)
	if err != nil {
		if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
			return err
		}
	}
	return nil
}

func createPersistentVolumeForRefDataset(client client.Client, virtualName string, virtualNameSpace string,
	runtime base.RuntimeInfoInterface, log logr.Logger) (err error) {

	pvName := getPvName(virtualName, virtualNameSpace)
	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		runtimePV, err := kubeclient.GetPersistentVolume(client, runtime.GetPersistentVolumeName())
		if err != nil {
			return err
		}

		runtimePVAttributes := runtimePV.Spec.PersistentVolumeSource.CSI.VolumeAttributes
		pv := &v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: virtualNameSpace,
				Labels: map[string]string{
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes:      runtimePV.Spec.AccessModes,
				Capacity:         runtimePV.Spec.Capacity.DeepCopy(),
				StorageClassName: common.FluidStorageClass,
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       common.CSIDriver,
						VolumeHandle: runtimePV.Spec.PersistentVolumeSource.CSI.VolumeHandle,
						VolumeAttributes: map[string]string{
							// 挂载点，使用runtime pv fuse path
							common.VolumeAttrFluidPath: runtimePVAttributes[common.VolumeAttrFluidPath],
							common.VolumeAttrMountType: runtimePVAttributes[common.VolumeAttrMountType],
							common.VolumeAttrNamespace: runtimePVAttributes[common.VolumeAttrNamespace],
							common.VolumeAttrName:      runtimePVAttributes[common.VolumeAttrName],
						},
					},
				},
				NodeAffinity: runtimePV.Spec.NodeAffinity.DeepCopy(),
			},
		}
		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}
	} else {
		log.Info("The ref persistent volume is created", "name", pvName)
	}

	return err
}

func createPersistentVolumeClaimForRefDataset(client client.Client, virtualName string, virtualNameSpace string,
	runtime base.RuntimeInfoInterface) (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, virtualName, virtualNameSpace, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		runtimePVC, err := kubeclient.GetPersistentVolumeClaim(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      virtualName,
				Namespace: virtualNameSpace,
				Labels: map[string]string{
					// see 'pkg/util/webhook/scheduler/mutating/schedule_pod_handler.go' 'CheckIfPVCIsDataset' function usage
					common.LabelAnnotationStorageCapacityPrefix + virtualNameSpace + "-" + virtualName: "true",
					common.LabelAnnotationDatasetReferringName:                                         runtimePVC.Name,
					common.LabelAnnotationDatasetReferringNameSpace:                                    runtimePVC.Namespace,
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						runtime.GetCommonLabelName(): "true",
					},
				},
				StorageClassName: &common.FluidStorageClass,
				AccessModes:      runtimePVC.Spec.AccessModes,
				Resources:        *runtimePVC.Spec.Resources.DeepCopy(),
			},
		}

		err = client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}

	return err
}
