package statefulset

import (
	"context"
	"fmt"
	apis "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/asts/apis"
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
	DeleteStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error
}

// AdvancedStatefulPodControl implements StatefulPodControlInterface using controller-runtime client.
type AdvancedStatefulPodControl struct {
	client   client.Client
	recorder record.EventRecorder
}

func NewAdvancedStatefulPodControl(client client.Client, recorder record.EventRecorder) AdvancedStatefulPodControlInterface {
	return &AdvancedStatefulPodControl{client: client, recorder: recorder}
}

func (spc *AdvancedStatefulPodControl) CreateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	// Create the Pod's PVCs prior to creating the Pod
	if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
		spc.recordPodEvent("create", set, pod, err)
		return err
	}
	// If we created the PVCs attempt to create the Pod
	err := spc.client.Create(context.TODO(), pod)
	// sink already exists errors
	if errors.IsAlreadyExists(err) {
		return err
	}
	spc.recordPodEvent("create", set, pod, err)
	return err
}

func (spc *AdvancedStatefulPodControl) UpdateStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
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
			if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
				spc.recordPodEvent("update", set, pod, err)
				return false, err
			}
		}
		// if the Pod is not dirty, do nothing
		if consistent {
			return true, nil
		}

		attemptedUpdate = true
		// commit the update, retrying on conflicts
		err := spc.client.Update(context.TODO(), pod)
		if err == nil {
			return true, nil
		}

		updatedPod := &corev1.Pod{}
		err = spc.client.Get(context.TODO(), client.ObjectKey{Namespace: set.Namespace, Name: pod.Name}, updatedPod)
		if err == nil {
			// make a copy so we don't mutate the shared cache
			pod = updatedPod.DeepCopy()
		} else {
			runtime.HandleError(fmt.Errorf("error getting updated Pod %s/%s from client: %v", set.Namespace, pod.Name, err))
		}

		return false, err
	})
	if attemptedUpdate {
		spc.recordPodEvent("update", set, pod, err)
	}
	return err
}

func (spc *AdvancedStatefulPodControl) DeleteStatefulPod(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	err := spc.client.Delete(context.TODO(), pod)
	spc.recordPodEvent("delete", set, pod, err)
	return err
}

// recordPodEvent records an event for verb applied to a Pod in a StatefulSet.
func (spc *AdvancedStatefulPodControl) recordPodEvent(verb string, set *apis.AdvancedStatefulSet, pod *corev1.Pod, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSet %s successful",
			strings.ToLower(verb), pod.Name, set.Name)
		spc.recorder.Event(set, corev1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSet %s failed error: %s",
			strings.ToLower(verb), pod.Name, set.Name, err)
		spc.recorder.Event(set, corev1.EventTypeWarning, reason, message)
	}
}

// createPersistentVolumeClaims creates all of the required PersistentVolumeClaims for pod.
func (spc *AdvancedStatefulPodControl) createPersistentVolumeClaims(set *apis.AdvancedStatefulSet, pod *corev1.Pod) error {
	var errs []error
	for _, claim := range getPersistentVolumeClaims(set, pod) {
		pvc := &corev1.PersistentVolumeClaim{}
		err := spc.client.Get(context.TODO(), client.ObjectKey{Namespace: claim.Namespace, Name: claim.Name}, pvc)
		switch {
		case errors.IsNotFound(err):
			err := spc.client.Create(context.TODO(), &claim)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create PVC %s: %s", claim.Name, err))
			}
			if err == nil || !errors.IsAlreadyExists(err) {
				spc.recordClaimEvent("create", set, pod, &claim, err)
			}
		case err != nil:
			errs = append(errs, fmt.Errorf("failed to retrieve PVC %s: %s", claim.Name, err))
			spc.recordClaimEvent("create", set, pod, &claim, err)
		}
		// TODO: Check resource requirements and accessmodes, update if necessary
	}
	return errors.NewAggregate(errs)

}

// recordClaimEvent records an event for verb applied to the PersistentVolumeClaim of a Pod in a StatefulSet.
func (spc *AdvancedStatefulPodControl) recordClaimEvent(verb string, set *apis.AdvancedStatefulSet, pod *corev1.Pod, claim *corev1.PersistentVolumeClaim, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s Pod %s in StatefulSet %s success",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name)
		spc.recorder.Event(set, corev1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s for Pod %s in StatefulSet %s failed error: %s",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name, err)
		spc.recorder.Event(set, corev1.EventTypeWarning, reason, message)
	}
}

var _ AdvancedStatefulPodControlInterface = &AdvancedStatefulPodControl{}
