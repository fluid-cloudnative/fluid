/*

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

package kubeclient

import (
	"context"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	persistentVolumeClaimProtectionFinalizerName = "kubernetes.io/pvc-protection"
)

func GetPersistentVolume(client client.Reader, name string) (pv *v1.PersistentVolume, err error) {
	pv = &v1.PersistentVolume{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: name}, pv)
	if err != nil {
		return nil, err
	}

	return pv, nil
}

// IsPersistentVolumeExist checks if the persistent volume exists given name and annotations of the PV.
func IsPersistentVolumeExist(client client.Client, name string, annotations map[string]string) (found bool, err error) {
	key := types.NamespacedName{
		Name: name,
	}

	pv := &v1.PersistentVolume{}

	err = client.Get(context.TODO(), key, pv)
	if err != nil {
		if apierrs.IsNotFound(err) {
			found = false
			err = nil
		}
	} else if len(pv.Annotations) == 0 {
		found = false
	} else {
		for k, v := range annotations {
			value := pv.Annotations[k]
			if value != v {
				found = false
				log.Info("The expected pv's annotation doesn't equal to what it has", "key", k,
					"expectedValue", v,
					"actualValue", value)
				return
			}
		}
		log.Info("The persistentVolume exist", "name", name,
			"annotaitons", annotations)
		found = true
	}

	return found, err
}

// IsPersistentVolumeClaimExist checks if the persistent volume claim exists given name, namespace and annotations of the PVC.
func IsPersistentVolumeClaimExist(client client.Client, name, namespace string, annotations map[string]string) (found bool, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	pvc := &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(), key, pvc)
	if err != nil {
		if apierrs.IsNotFound(err) {
			found = false
			err = nil
		}
	} else if len(pvc.Annotations) == 0 {
		found = false
	} else {
		for k, v := range annotations {
			value := pvc.Annotations[k]
			if value != v {
				found = false
				log.Info("The expected pvc's annotation doesn't equal to what it has", "key", k,
					"expectedValue", v,
					"actualValue", value)
				return
			}
		}
		log.Info("The persistentVolume exist", "name", name,
			"annotaitons", annotations)
		found = true
	}

	return found, err

}

// DeletePersistentVolume deletes volume
func DeletePersistentVolume(client client.Client, name string) (err error) {
	key := types.NamespacedName{
		Name: name,
	}
	found := false
	pv := &v1.PersistentVolume{}
	if err = client.Get(context.TODO(), key, pv); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the PersistentVolume due to it's not found", "name", name)
			found = false
			err = nil
		} else {
			return
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), pv)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("error deleting pv %s: %s", name, err.Error())
		}
	}

	return
}

// DeletePersistentVolumeClaim deletes volume claim
func DeletePersistentVolumeClaim(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	found := false
	pvc := &v1.PersistentVolumeClaim{}
	if err = client.Get(context.TODO(), key, pvc); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleting the PersistentVolumeClaim due to it's not found", "name", name,
				"namespace", namespace)
			found = false
			err = nil
		} else {
			return
		}
	} else {
		found = true
	}
	if found {
		log.V(1).Info("deleting pvc", "PVC", pvc)
		err = client.Delete(context.TODO(), pvc)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("error deleting pvc %s: %s", name, err.Error())
		}
	}

	return
}

// GetPVCsFromPod get PersistVolumeClaims of pod
func GetPVCsFromPod(pod v1.Pod) (pvcs []v1.Volume) {
	for _, volume := range pod.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			pvcs = append(pvcs, volume)
		}
	}
	return pvcs
}

// GetPvcMountPods get pods that mounted the specific pvc for a given namespace
func GetPvcMountPods(e client.Client, pvcName, namespace string) ([]v1.Pod, error) {
	nsPods := v1.PodList{}
	err := e.List(context.TODO(), &nsPods, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		log.Error(err, "Failed to list pods")
		return []v1.Pod{}, err
	}
	var pods []v1.Pod
	for _, pod := range nsPods.Items {
		pvcs := GetPVCsFromPod(pod)
		for _, pvc := range pvcs {
			if pvc.PersistentVolumeClaim.ClaimName == pvcName {
				pods = append(pods, pod)
			}
		}
	}
	return pods, err
}

// GetPvcMountNodes get nodes which have pods mounted the specific pvc for a given namespace
// it will only return a map of nodeName and amount of PvcMountPods on it
// if the Pvc mount Pod has completed, it will be ignored
// if fail to get pvc mount Nodes, treat every nodes as with no PVC mount Pods
func GetPvcMountNodes(e client.Client, pvcName, namespace string) (map[string]int64, error) {
	pvcMountNodes := map[string]int64{}
	pvcMountPods, err := GetPvcMountPods(e, pvcName, namespace)
	if err != nil {
		log.Error(err, "Failed to get PVC Mount Nodes because cannot list pods")
		return pvcMountNodes, err
	}

	for _, pod := range pvcMountPods {
		if IsCompletePod(&pod) {
			continue
		}
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}
		if _, found := pvcMountNodes[nodeName]; !found {
			pvcMountNodes[nodeName] = 1
		} else {
			pvcMountNodes[nodeName] = pvcMountNodes[nodeName] + 1
		}
	}
	return pvcMountNodes, nil
}

// RemoveProtectionFinalizer remove finalizers of PersistentVolumeClaim
// if all owners that this PVC is mounted by are inactive (Succeed or Failed)
func RemoveProtectionFinalizer(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	pvc := &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(), key, pvc)
	if err != nil && !apierrs.IsNotFound(err) {
		return err
	}

	// try to remove finalizer "pvc-protection"
	log.Info("Remove finalizer pvc-protection")
	finalizers := utils.RemoveString(pvc.Finalizers, persistentVolumeClaimProtectionFinalizerName)
	pvc.SetFinalizers(finalizers)
	if err = client.Update(context.TODO(), pvc); err != nil {
		log.Error(err, "Failed to remove finalizer",
			"Finalizer", persistentVolumeClaimProtectionFinalizerName)
		return err
	}

	return err
}

// ShouldDeleteDataset return no err when no pod is using the volume
// If cannot get PVC, cannot get PvcMountPods, or running pod is using the volume, return corresponding error
func ShouldDeleteDataset(client client.Client, name, namespace string) (err error) {
	// 1. Check if the pvc exists
	exist, err := IsPersistentVolumeClaimExist(client, name, namespace, common.ExpectedFluidAnnotations)
	if err != nil {
		return
	}
	if !exist {
		return nil
	}

	// 2. check if the pod on it is running
	pods, err := GetPvcMountPods(client, name, namespace)
	if err != nil {
		return
	}
	for _, pod := range pods {
		if !IsCompletePod(&pod) {
			err = fmt.Errorf("can not delete dataset "+
				"because Pod %v in Namespace %v is using it", pod.Name, pod.Namespace)
			return
		}
	}
	return nil
}

// ShouldRemoveProtectionFinalizer should remove pvc-protection finalizer
// when linked pods are inactive and timeout
func ShouldRemoveProtectionFinalizer(client client.Client, name, namespace string) (should bool, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	pvc := &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(), key, pvc)
	if err != nil {
		return
	}

	if pvc.DeletionTimestamp.IsZero() ||
		!utils.ContainsString(pvc.Finalizers, persistentVolumeClaimProtectionFinalizerName) {
		return
	}

	// only force remove finalizer after 30 seconds' Terminating state
	then := pvc.DeletionTimestamp.Add(30 * time.Second)
	now := time.Now()
	if now.Before(then) {
		log.V(1).Info("can not remove pvc-protection finalizer before reached expected timeout",
			"Remaining seconds:", then.Sub(now).Seconds())
		return
	}

	// get pods which mounted this pvc
	pods, err := GetPvcMountPods(client, name, namespace)
	if err != nil {
		return
	}
	// check pods status
	for _, pod := range pods {
		if !IsCompletePod(&pod) {
			err = fmt.Errorf("can not remove pvc-protection finalizer "+
				"because Pod %v in Namespace %v is not completed", pod.Name, pod.Namespace)
			return
		}
	}

	should = true

	return
}

// IsDatasetPVC check whether the PVC is a dataset PVC
func IsDatasetPVC(client client.Reader, name string, namespace string) (find bool, err error) {
	pvc := &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, pvc)
	if err != nil {
		return
	}
	_, find = pvc.Labels[common.LabelAnnotationStorageCapacityPrefix+namespace+"-"+name]
	return
}
