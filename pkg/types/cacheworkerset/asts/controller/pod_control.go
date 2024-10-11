package statefulset

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	apis "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/asts/apis"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// StatefulPodControlInterface defines the interface for stateful pod control.
type AdvancedStatefulPodControlInterface interface {
	CreateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error
	UpdateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error
	DeleteStatefulPodAndCleanUpCache(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error
}

// AdvancedStatefulPodControl implements StatefulPodControlInterface using controller-runtime client.
type AdvancedStatefulPodControl struct {
	client   client.Client
	recorder record.EventRecorder
}

func NewAdvancedStatefulPodControl(client client.Client, recorder record.EventRecorder) AdvancedStatefulPodControlInterface {
	return &AdvancedStatefulPodControl{client: client, recorder: recorder}
}

func (asts *AdvancedStatefulPodControl) CreateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	// Create the Pod's PVCs prior to creating the Pod
	if err := asts.createPersistentVolumeClaims(set, pod); err != nil {
		asts.recordPodEvent("create", set, pod, err)
		return err
	}
	// If we created the PVCs attempt to create the Pod
	err := asts.client.Create(context.TODO(), pod)
	// sink already exists errors
	if errors.IsAlreadyExists(err) {
		return err
	}
	asts.recordPodEvent("create", set, pod, err)
	return err
}

func (asts *AdvancedStatefulPodControl) UpdateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	attemptedUpdate := false
	err := wait.ExponentialBackoff(wait.Backoff{
		Steps:    4,
		Duration: 100 * wait.Millisecond,
		Factor:   1.5,
		Jitter:   0.1,
	}, func() (bool, error) {
		// assume the Pod is consistent
		consistent := true
		// if the Pod does not conform to its identity, update the identity and dirty the Pod
		if !identityMatches(set, pod) {
			updateIdentity(set, pod)
			consistent = false
		}
		// if the Pod does not conform to the StatefulSet's storage requirements, update the Pod's PVC's,
		// dirty the Pod, and create any missing PVCs
		if !storageMatches(set, pod) {
			updateStorage(set, pod)
			consistent = false
			if err := asts.createPersistentVolumeClaims(set, pod); err != nil {
				asts.recordPodEvent("update", set, pod, err)
				return false, err
			}
		}
		// if the Pod is not dirty, do nothing
		if consistent {
			return true, nil
		}

		attemptedUpdate = true
		// commit the update, retrying on conflicts
		err := asts.client.Update(context.TODO(), pod)
		if err == nil {
			return true, nil
		}

		updatedPod := &corev1.Pod{}
		err = asts.client.Get(context.TODO(), client.ObjectKey{Namespace: set.Namespace, Name: pod.Name}, updatedPod)
		if err == nil {
			// make a copy so we don't mutate the shared cache
			pod = updatedPod.DeepCopy()
		} else {
			runtime.HandleError(fmt.Errorf("error getting updated Pod %s/%s from client: %v", set.Namespace, pod.Name, err))
		}

		return false, err
	})
	if attemptedUpdate {
		asts.recordPodEvent("update", set, pod, err)
	}
	return err
}

//
//func (asts *AdvancedStatefulPodControl) DeleteStatefulPodAndCleanUpCache(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
//	// 获取 AlluxioEngine 实例（假设已经有正确的获取方式）
//	engine := alluxio.AlluxioEngine{}
//
//	// 清理与该 pod 相关的缓存
//	err := engine.cleanupCacheForPod(pod)
//	if err != nil {
//		asts.recordPodEvent("delete", set, pod, fmt.Errorf("failed to clean cache: %v", err))
//		return fmt.Errorf("failed to clean cache: %v", err)
//	}
//
//	// 释放与该 pod 相关的端口（如果需要）
//	err = engine.releasePortsForPod(pod)
//	if err != nil {
//		asts.recordPodEvent("delete", set, pod, fmt.Errorf("failed to release ports: %v", err))
//		return fmt.Errorf("failed to release ports: %v", err)
//	}
//
//	// 删除该 pod
//	err = asts.client.Delete(context.TODO(), pod)
//	if err != nil {
//		asts.recordPodEvent("delete", set, pod, fmt.Errorf("failed to delete pod: %v", err))
//		return fmt.Errorf("failed to delete pod: %v", err)
//	}
//
//	// 删除与该 pod 相关的数据集（假设根据标签查找和删除）
//	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", someDatasetLabelName))
//	if err != nil {
//		return err
//	}
//	datasetList := &someDatasetTypeList{}
//	err = asts.client.List(context.TODO(), datasetList, &client.ListOptions{LabelSelector: datasetLabels})
//	if err != nil {
//		return err
//	}
//	for _, dataset := range datasetList.Items {
//		err = asts.client.Delete(context.TODO(), &dataset)
//		if err != nil {
//			asts.recordPodEvent("delete", set, pod, fmt.Errorf("failed to delete dataset: %v", err))
//			return fmt.Errorf("failed to delete dataset: %v", err)
//		}
//	}
//
//	asts.recordPodEvent("delete", set, pod, nil)
//	return nil
//}

// DestroySpecificWorkerPod destroys a specific worker pod and performs related cleanup operations.
func (asts *AdvancedStatefulPodControl) DeleteStatefulPodAndCleanUpCache(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	var alluxio_runtime *datav1alpha1.AlluxioRuntime
	e := alluxio.AlluxioEngine{
		Runtime:            alluxio_runtime,
		Log:                logr.Logger{},
		Client:             nil,
		MetadataSyncDoneCh: nil,
		Name:               "a",
		Namespace:          "b",
		UnitTest:           false,
		Helper:             nil,
		Recorder:           nil,
	}
	// Delete the specific worker pod.
	err := e.DestroyPodsByAnnotationOnWorkerNodes(deletePodsAnnoKey, deletePodsAnnoValue)
	if err != nil {
		asts.recordPodEvent("delete", set, pod, fmt.Errorf("failed to delete pod: %v", err))
		return fmt.Errorf("failed to delete pod: %v", err)
	}

	if err != nil {
		return err
	}

	// Delete the corresponding dataset if needed.
	// Implement logic to determine if the dataset should be deleted and perform the deletion.

	return nil
}

// recordPodEvent records an event for verb applied to a Pod in a StatefulSet.
func (asts *AdvancedStatefulPodControl) recordPodEvent(verb string, set *apis.AdvancedStatefulSet, pod *corev1.Pod, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSet %s successful",
			strings.ToLower(verb), pod.Name, set.Name)
		asts.recorder.Event(set, corev1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSet %s failed error: %s",
			strings.ToLower(verb), pod.Name, set.Name, err)
		asts.recorder.Event(set, corev1.EventTypeWarning, reason, message)
	}
}

// createPersistentVolumeClaims creates all of the required PersistentVolumeClaims for pod.
func (asts *AdvancedStatefulPodControl) createPersistentVolumeClaims(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	var errs []error
	for _, claim := range getPersistentVolumeClaims(set, pod) {
		pvc := &corev1.PersistentVolumeClaim{}
		err := asts.client.Get(context.TODO(), client.ObjectKey{Namespace: claim.Namespace, Name: claim.Name}, pvc)
		switch {
		case errors.IsNotFound(err):
			err := asts.client.Create(context.TODO(), &claim)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create PVC %s: %s", claim.Name, err))
			}
			if err == nil || !errors.IsAlreadyExists(err) {
				asts.recordClaimEvent("create", set, pod, &claim, err)
			}
		case err != nil:
			errs = append(errs, fmt.Errorf("failed to retrieve PVC %s: %s", claim.Name, err))
			asts.recordClaimEvent("create", set, pod, &claim, err)
		}
		// TODO: Check resource requirements and accessmodes, update if necessary
	}
	return errors.NewAggregate(errs)

}

// recordClaimEvent records an event for verb applied to the PersistentVolumeClaim of a Pod in a StatefulSet.
func (asts *AdvancedStatefulPodControl) recordClaimEvent(verb string, set *apis.AdvancedStatefulSet, pod *corev1.Pod, claim *corev1.PersistentVolumeClaim, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s Pod %s in StatefulSet %s success",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name)
		asts.recorder.Event(set, corev1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s for Pod %s in StatefulSet %s failed error: %s",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name, err)
		asts.recorder.Event(set, corev1.EventTypeWarning, reason, message)
	}
}

var _ AdvancedStatefulPodControlInterface = &AdvancedStatefulPodControl{}
