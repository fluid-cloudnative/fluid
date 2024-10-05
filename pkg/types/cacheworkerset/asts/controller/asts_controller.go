package statefulset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	apis "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/asts/apis"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"k8s.io/client-go/kubernetes/scheme"
	clientsetappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"math"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
	"strconv"
	"strings"
)

// StatefulSetReconciler reconciles a StatefulSet object.
type StatefulSetReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
	client     client.Client
	Log        logr.Logger
	PodControl AdvancedStatefulPodControl
	Recorder   record.EventRecorder
}

// Reconcile is the main reconciliation loop for StatefulSet.
func (r *StatefulSetReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	//namespace, name, err := cache.SplitMetaNamespaceKey(req.NamespacedName.String())
	//if err != nil {
	//	return reconcile.Result{}, err
	//}
	set := &apis.AdvancedStatefulSet{}
	err := r.Client.Get(ctx, req.NamespacedName, set)
	if errors.IsNotFound(err) {
		logger.Info("StatefulSet has been deleted")
		return reconcile.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "unable to retrieve StatefulSet")
		return reconcile.Result{}, err
	}

	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		logger.Error(err, "error converting StatefulSet selector")
		return reconcile.Result{}, nil
	}

	revisions, err := r.ListRevisions(set)
	if err != nil {
		logger.Error(err, "error listing ControllerRevisions")
		return reconcile.Result{}, err
	}

	currentRevision, updateRevision, collisionCount, err := r.getStatefulSetRevisions(set, revisions)
	if err != nil {
		logger.Error(err, "error getting StatefulSet revisions")
		return reconcile.Result{}, err
	}

	pods, err := r.getPodsForStatefulSet(set, selector)
	if err != nil {
		logger.Error(err, "error getting pods for StatefulSet")
		return reconcile.Result{}, err
	}

	status, err := r.updateStatefulSet(set, currentRevision, updateRevision, collisionCount, pods)
	if err != nil {
		logger.Error(err, "error updating StatefulSet")
		return reconcile.Result{}, err
	}

	err = r.updateStatefulSetStatus(set, status)
	if err != nil {
		logger.Error(err, "error updating StatefulSet status")
		return reconcile.Result{}, err
	}

	logger.Info("Successfully synced StatefulSet", "namespace", set.Namespace, "name", set.Name)

	return reconcile.Result{}, nil
}

// ListRevisions returns a array of the ControllerRevisions that represent the revisions of set.
func (r *StatefulSetReconciler) ListRevisions(set *apis.AdvancedStatefulSet) ([]*appsv1.ControllerRevision, error) {
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		return nil, err
	}
	revisions := &appsv1.ControllerRevisionList{}
	err = r.Client.List(context.TODO(), revisions, client.InNamespace(set.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	kubeClient, ok := r.client.(kubernetes.Interface)
	if !ok {
		return nil, fmt.Errorf("could not convert client to kubernetes.Interface")
	}
	appsV1Client := clientsetappsv1.AppsV1Interface(kubeClient.AppsV1())

	revisionsToUpgrade, err := appsV1Client.ControllerRevisions(set.GetNamespace()).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
				apis.UpgradeToAdvancedStatefulSetAnn: set.Name,
			}).String(),
		})
	if err != nil {
		return nil, err
	}

	res := []*appsv1.ControllerRevision{}
	for _, item := range append(revisions.Items, revisionsToUpgrade.Items...) {
		local := item
		res = append(res, &local)
	}
	return res, nil
}

func (r *StatefulSetReconciler) updateControllerRevision(revision *appsv1.ControllerRevision, newRevision int64) (*appsv1.ControllerRevision, error) {
	clone := revision.DeepCopy()
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if clone.Revision == newRevision {
			return nil
		}
		clone.Revision = newRevision
		kubeClient, ok := r.client.(kubernetes.Interface)
		if !ok {
			return fmt.Errorf("could not convert client to kubernetes.Interface")
		}
		appsV1Client := clientsetappsv1.AppsV1Interface(kubeClient.AppsV1())
		updated, updateErr := appsV1Client.ControllerRevisions(clone.Namespace).Update(context.TODO(), clone, metav1.UpdateOptions{})
		if updateErr == nil {
			return nil
		}
		if updated != nil {
			clone = updated
		}

		if updated, err := appsV1Client.ControllerRevisions(clone.Namespace).Get(context.TODO(), clone.Name, metav1.GetOptions{}); err == nil {
			// make a copy so we don't mutate the shared cache
			clone = updated.DeepCopy()
		}
		return updateErr
	})
	return clone, err
}

func (r *StatefulSetReconciler) createControllerRevision(parent metav1.Object, revision *appsv1.ControllerRevision, collisionCount *int32) (*appsv1.ControllerRevision, error) {
	if collisionCount == nil {
		return nil, fmt.Errorf("collisionCount should not be nil")
	}

	// Clone the input
	clone := revision.DeepCopy()

	// Continue to attempt to create the revision updating the name with a new hash on each iteration
	for {
		hash := hashControllerRevision(revision, collisionCount)
		// Update the revisions name
		clone.Name = controllerRevisionName(parent.GetName(), hash)
		ns := parent.GetNamespace()
		kubeClient, ok := r.client.(kubernetes.Interface)
		if !ok {
			return nil, fmt.Errorf("could not convert client to kubernetes.Interface")
		}
		appsV1Client := clientsetappsv1.AppsV1Interface(kubeClient.AppsV1())
		created, err := appsV1Client.ControllerRevisions(ns).Create(context.TODO(), clone, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			exists, err := appsV1Client.ControllerRevisions(ns).Get(context.TODO(), clone.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if bytes.Equal(exists.Data.Raw, clone.Data.Raw) {
				return exists, nil
			}
			*collisionCount++
			continue
		}
		return created, err
	}
}

// getStatefulSetRevisions returns the current and update ControllerRevisions for set.
func (r *StatefulSetReconciler) getStatefulSetRevisions(set *apis.AdvancedStatefulSet, revisions []*appsv1.ControllerRevision) (*appsv1.ControllerRevision, *appsv1.ControllerRevision, int32, error) {
	var currentRevision, updateRevision *appsv1.ControllerRevision

	revisionCount := len(revisions)
	sort.Stable(byRevision(revisions))

	var collisionCount int32
	if set.Status.CollisionCount != nil {
		collisionCount = *set.Status.CollisionCount
	}

	updateRevision, err := newRevision(set, nextRevision(revisions), &collisionCount)
	if err != nil {
		return nil, nil, collisionCount, err
	}

	equalRevisions := findEqualRevisions(revisions, updateRevision)
	equalCount := len(equalRevisions)

	if equalCount > 0 && equalRevision(revisions[revisionCount-1], equalRevisions[equalCount-1]) {
		updateRevision = revisions[revisionCount-1]
	} else if equalCount > 0 {
		updateRevision, err = r.updateControllerRevision(equalRevisions[equalCount-1], updateRevision.Revision)
		if err != nil {
			return nil, nil, collisionCount, err
		}
	} else {
		updateRevision, err = r.createControllerRevision(set, updateRevision, &collisionCount)
		if err != nil {
			return nil, nil, collisionCount, err
		}
	}

	for i := range revisions {
		if revisions[i].Name == set.Status.CurrentRevision {
			currentRevision = revisions[i]
			break
		}
	}

	if currentRevision == nil {
		currentRevision = updateRevision
	}

	return currentRevision, updateRevision, collisionCount, nil
}

// updateStatefulSet performs the update function for a StatefulSet.
func (r *StatefulSetReconciler) updateStatefulSet(set *apis.AdvancedStatefulSet, currentRevision *appsv1.ControllerRevision, updateRevision *appsv1.ControllerRevision, collisionCount int32, pods []*corev1.Pod) (*apis.AdvancedStatefulSetStatus, error) {
	currentSet, err := applyRevision(set, currentRevision)
	if err != nil {
		return nil, err
	}
	updateSet, err := applyRevision(set, updateRevision)
	if err != nil {
		return nil, err
	}

	status := apis.AdvancedStatefulSetStatus{}
	status.ObservedGeneration = set.Generation
	status.CurrentRevision = currentRevision.Name
	status.UpdateRevision = updateRevision.Name
	status.CollisionCount = new(int32)
	*status.CollisionCount = collisionCount

	deleteSlots := apis.GetDeleteSlots(set)
	_replicaCount, deleteSlots := apis.GetMaxReplicaCountAndDeleteSlots(*set.Spec.Replicas, deleteSlots)
	replicaCount := int(_replicaCount)

	replicas := make([]*corev1.Pod, replicaCount)
	condemned := make([]*corev1.Pod, 0, len(pods))
	unhealthy := 0
	firstUnhealthyOrdinal := math.MaxInt32
	var firstUnhealthyPod *corev1.Pod

	for i := range pods {
		status.Replicas++

		if isRunningAndReady(pods[i]) {
			status.ReadyReplicas++
		}

		if isCreated(pods[i]) && !isTerminating(pods[i]) {
			if getPodRevision(pods[i]) == currentRevision.Name {
				status.CurrentReplicas++
			}
			if getPodRevision(pods[i]) == updateRevision.Name {
				status.UpdatedReplicas++
			}
		}

		ord := getOrdinal(pods[i])
		if 0 <= ord && ord < replicaCount && !deleteSlots.Has(int32(ord)) {
			replicas[ord] = pods[i]
		} else if ord >= replicaCount || deleteSlots.Has(int32(ord)) {
			condemned = append(condemned, pods[i])
		}
	}

	for ord := 0; ord < replicaCount; ord++ {
		if deleteSlots.Has(int32(ord)) {
			continue
		}
		if replicas[ord] == nil {
			replicas[ord] = newVersionedStatefulSetPod(currentSet, updateSet, currentRevision.Name, updateRevision.Name, ord)
		}
	}

	sort.Sort(ascendingOrdinal(condemned))

	for i := range replicas {
		if replicas[i] == nil {
			continue
		}
		if !isHealthy(replicas[i]) {
			unhealthy++
			if ord := getOrdinal(replicas[i]); ord < firstUnhealthyOrdinal {
				firstUnhealthyOrdinal = ord
				firstUnhealthyPod = replicas[i]
			}
		}
	}

	for i := range condemned {
		if !isHealthy(condemned[i]) {
			unhealthy++
			if ord := getOrdinal(condemned[i]); ord < firstUnhealthyOrdinal {
				firstUnhealthyOrdinal = ord
				firstUnhealthyPod = condemned[i]
			}
		}
	}

	if unhealthy > 0 {
		log.Log.Info("StatefulSet has unhealthy pods", "namespace", set.Namespace, "name", set.Name, "unhealthyCount", unhealthy, "firstUnhealthyPod", firstUnhealthyPod.Name)
	}

	if set.DeletionTimestamp != nil {
		return &status, nil
	}

	monotonic := !allowsBurst(set)

	for i := range replicas {
		if replicas[i] == nil {
			continue
		}
		if isFailed(replicas[i]) || isSucceeded(replicas[i]) {
			r.Recorder.Eventf(set, corev1.EventTypeWarning, "RecreatingFailedPod", "StatefulSet %s/%s is recreating failed Pod %s", set.Namespace, set.Name, replicas[i].Name)
			if err := r.PodControl.DeleteStatefulPod(set, replicas[i]); err != nil {
				return &status, err
			}
			if getPodRevision(replicas[i]) == currentRevision.Name {
				status.CurrentReplicas--
			}
			if getPodRevision(replicas[i]) == updateRevision.Name {
				status.UpdatedReplicas--
			}
			status.Replicas--
			replicas[i] = newVersionedStatefulSetPod(currentSet, updateSet, currentRevision.Name, updateRevision.Name, i)
		}
		if !isCreated(replicas[i]) {
			if err := r.PodControl.CreateStatefulPod(set, replicas[i]); err != nil {
				return &status, err
			}
			status.Replicas++
			if getPodRevision(replicas[i]) == currentRevision.Name {
				status.CurrentReplicas++
			}
			if getPodRevision(replicas[i]) == updateRevision.Name {
				status.UpdatedReplicas++
			}

			if monotonic {
				return &status, nil
			}
			continue
		}
		if isTerminating(replicas[i]) && monotonic {
			log.Log.Info("Waiting for pod to terminate", "namespace", set.Namespace, "name", set.Name, "podName", replicas[i].Name)
			return &status, nil
		}
		if !isRunningAndReady(replicas[i]) && monotonic {
			log.Log.Info("Waiting for pod to be running and ready", "namespace", set.Namespace, "name", set.Name, "podName", replicas[i].Name)
			return &status, nil
		}
		if identityMatches(set, replicas[i]) && storageMatches(set, replicas[i]) {
			continue
		}
		replica := replicas[i].DeepCopy()
		if err := r.PodControl.UpdateStatefulPod(updateSet, replica); err != nil {
			return &status, err
		}
	}

	for target := len(condemned) - 1; target >= 0; target-- {
		if isTerminating(condemned[target]) {
			log.Log.Info("Waiting for pod to terminate prior to scale down", "namespace", set.Namespace, "name", set.Name, "podName", condemned[target].Name)
			if monotonic {
				return &status, nil
			}
			continue
		}
		if !isRunningAndReady(condemned[target]) && monotonic && condemned[target] != firstUnhealthyPod {
			log.Log.Info("Waiting for pod to be running and ready prior to scale down", "namespace", set.Namespace, "name", set.Name, "podName", firstUnhealthyPod.Name)
			return &status, nil
		}
		log.Log.Info("Terminating pod for scale down", "namespace", set.Namespace, "name", set.Name, "podName", condemned[target].Name)

		if err := r.PodControl.DeleteStatefulPod(set, condemned[target]); err != nil {
			return &status, err
		}
		if getPodRevision(condemned[target]) == currentRevision.Name {
			status.CurrentReplicas--
		}
		if getPodRevision(condemned[target]) == updateRevision.Name {
			status.UpdatedReplicas--
		}
		if monotonic {
			return &status, nil
		}
	}

	if set.Spec.UpdateStrategy.Type == apis.OnDeleteStatefulSetStrategyType {
		return &status, nil
	}

	updateMin := 0
	if set.Spec.UpdateStrategy.RollingUpdate != nil {
		updateMin = int(*set.Spec.UpdateStrategy.RollingUpdate.Partition)
	}
	for target := len(replicas) - 1; target >= updateMin; target-- {
		if replicas[target] == nil {
			continue
		}
		if getPodRevision(replicas[target]) != updateRevision.Name && !isTerminating(replicas[target]) {
			log.Log.Info("Terminating pod for update", "namespace", set.Namespace, "name", set.Name, "podName", replicas[target].Name)
			err := r.PodControl.DeleteStatefulPod(set, replicas[target])
			status.CurrentReplicas--
			return &status, err
		}

		if !isHealthy(replicas[target]) {
			log.Log.Info("Waiting for pod to update", "namespace", set.Namespace, "name", set.Name, "podName", replicas[target].Name)
			return &status, nil
		}

	}
	return &status, nil
}

// updateStatefulSetStatus updates set's Status to be equal to status.
func (r *StatefulSetReconciler) updateStatefulSetStatus(set *apis.AdvancedStatefulSet, status *apis.AdvancedStatefulSetStatus) error {
	completeRollingUpdate(set, status)

	if !inconsistentStatus(set, status) {
		return nil
	}

	//set = set.DeepCopy()，这个地方存在争议，？暂时写不出来
	return r.Client.Status().Update(context.TODO(), set)
}

//// podControl is used for patching pods.
//type podControl struct {
//	Client   client.Client
//	Recorder record.EventRecorder
//}

// recordClaimEvent records an event for verb applied to the PersistentVolumeClaim of a Pod in a StatefulSet. If err is
// nil the generated event will have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a
// reason of v1.EventTypeWarning.
func (r *StatefulSetReconciler) recordClaimEvent(verb string, set *apis.AdvancedStatefulSet, pod *corev1.Pod, claim *corev1.PersistentVolumeClaim, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s Pod %s in StatefulSet %s success",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name)
		r.Recorder.Event(set, corev1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s for Pod %s in StatefulSet %s failed error: %s",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name, err)
		r.Recorder.Event(set, corev1.EventTypeWarning, reason, message)
	}
}

// getPodsForStatefulSet returns the pods associated with a StatefulSet.
func (r *StatefulSetReconciler) getPodsForStatefulSet(set *apis.AdvancedStatefulSet, selector labels.Selector) ([]*corev1.Pod, error) {
	podList := &corev1.PodList{}
	err := r.Client.List(context.TODO(), podList, client.InNamespace(set.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	pods := make([]*corev1.Pod, 0, len(podList.Items))
	for i := range podList.Items {
		pods = append(pods, &podList.Items[i])
	}
	return pods, nil
}

// SetupWithManager setups the manager with RuntimeReconciler
func (r *StatefulSetReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.AlluxioRuntime{}).
		Complete(r)
}

// newRevision creates a new ControllerRevision for a StatefulSet.
func newRevision(set *apis.AdvancedStatefulSet, revision *appsv1.ControllerRevision, collisionCount *int32) (*appsv1.ControllerRevision, error) {
	patch, err := getPatch(set)
	if err != nil {
		return nil, err
	}
	var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v2"}
	controllerKind := SchemeGroupVersion.WithKind("AdvancedStatefulset")
	cr, err := NewControllerRevision(set,
		controllerKind,
		set.Spec.Template.Labels,
		runtime.RawExtension{Raw: patch},
		revision.Revision,
		collisionCount)
	if err != nil {
		return nil, err
	}
	if cr.ObjectMeta.Annotations == nil {
		cr.ObjectMeta.Annotations = make(map[string]string)
	}
	for key, value := range set.Annotations {
		cr.ObjectMeta.Annotations[key] = value
	}
	return cr, nil
}

// findEqualRevisions finds ControllerRevisions that are equal to the given revision.
func findEqualRevisions(revisions []*appsv1.ControllerRevision, revision *appsv1.ControllerRevision) []*appsv1.ControllerRevision {
	equalRevisions := make([]*appsv1.ControllerRevision, 0)
	for _, r := range revisions {
		if equalRevision(r, revision) {
			equalRevisions = append(equalRevisions, r)
		}
	}
	return equalRevisions
}

// equalRevision checks if two ControllerRevisions are equal.
func equalRevision(a, b *appsv1.ControllerRevision) bool {
	return bytes.Equal(a.Data.Raw, b.Data.Raw)
}

// applyRevision applies a ControllerRevision to a StatefulSet.
func applyRevision(set *apis.AdvancedStatefulSet, revision *appsv1.ControllerRevision) (*apis.AdvancedStatefulSet, error) {
	clone := set.DeepCopy()
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(scheme.Codecs, clone)), revision.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	restoredSet := &apis.AdvancedStatefulSet{}
	err = json.Unmarshal(patched, restoredSet)
	if err != nil {
		return nil, err
	}
	return restoredSet, nil
}

// nextRevision finds the next valid revision number based on revisions. If the length of revisions
// is 0 this is 1. Otherwise, it is 1 greater than the largest revision's Revision. This method
// assumes that revisions has been sorted by Revision.
func nextRevision(revisions []*appsv1.ControllerRevision) int64 {
	count := len(revisions)
	if count <= 0 {
		return 1
	}
	return revisions[count-1].Revision + 1
}

// inconsistentStatus checks if a StatefulSet's status is inconsistent.
func inconsistentStatus(set *apis.AdvancedStatefulSet, status *apis.AdvancedStatefulSetStatus) bool {
	return status.ObservedGeneration > set.Status.ObservedGeneration ||
		status.Replicas != set.Status.Replicas ||
		status.CurrentReplicas != set.Status.CurrentReplicas ||
		status.ReadyReplicas != set.Status.ReadyReplicas ||
		status.UpdatedReplicas != set.Status.UpdatedReplicas ||
		status.CurrentRevision != set.Status.CurrentRevision ||
		status.UpdateRevision != set.Status.UpdateRevision
}

// completeRollingUpdate completes a rolling update for a StatefulSet.
func completeRollingUpdate(set *apis.AdvancedStatefulSet, status *apis.AdvancedStatefulSetStatus) {
	if set.Spec.UpdateStrategy.Type == apis.RollingUpdateStatefulSetStrategyType &&
		status.UpdatedReplicas == status.Replicas &&
		status.ReadyReplicas == status.Replicas {
		status.CurrentReplicas = status.UpdatedReplicas
		status.CurrentRevision = status.UpdateRevision
	}
}

// overlappingStatefulSets sorts StatefulSets by creation timestamp and name.
type overlappingStatefulSets []*apis.AdvancedStatefulSet

func (o overlappingStatefulSets) Len() int {
	return len(o)
}

func (o overlappingStatefulSets) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o overlappingStatefulSets) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// statefulPodRegex is a regular expression to extract parent StatefulSet and ordinal from a Pod name.
var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the parent StatefulSet name and ordinal from a Pod.
func getParentNameAndOrdinal(pod *corev1.Pod) (string, int) {
	parent := ""
	ordinal := -1
	subMatches := statefulPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return parent, ordinal
	}
	parent = subMatches[1]
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return parent, ordinal
}

// getParentName gets the parent StatefulSet name from a Pod.
func getParentName(pod *corev1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

// byRevision implements sort.Interface to allow ControllerRevisions to be sorted by Revision.
type byRevision []*appsv1.ControllerRevision

func (br byRevision) Len() int {
	return len(br)
}

// Less breaks ties first by creation timestamp, then by name
func (br byRevision) Less(i, j int) bool {
	if br[i].Revision == br[j].Revision {
		if br[j].CreationTimestamp.Equal(&br[i].CreationTimestamp) {
			return br[i].Name < br[j].Name
		}
		return br[j].CreationTimestamp.After(br[i].CreationTimestamp.Time)
	}
	return br[i].Revision < br[j].Revision
}

func (br byRevision) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}

// getOrdinal gets the ordinal from a Pod.
func getOrdinal(pod *corev1.Pod) int {
	_, ordinal := getParentNameAndOrdinal(pod)
	return ordinal
}

// getPodName gets the name of a Pod for a StatefulSet and ordinal.
func getPodName(set *apis.AdvancedStatefulSet, ordinal int) string {
	return fmt.Sprintf("%s-%d", set.Name, ordinal)
}

// getPersistentVolumeClaimName gets the name of a PersistentVolumeClaim for a Pod.
func getPersistentVolumeClaimName(set *apis.AdvancedStatefulSet, claim *corev1.PersistentVolumeClaim, ordinal int) string {
	return fmt.Sprintf("%s-%s-%d", claim.Name, set.Name, ordinal)
}

// isMemberOf checks if a Pod is a member of a StatefulSet.
func isMemberOf(set *apis.AdvancedStatefulSet, pod *corev1.Pod) bool {
	return getParentName(pod) == set.Name
}

// identityMatches checks if a Pod has a valid identity for a StatefulSet.
func identityMatches(set *apis.AdvancedStatefulSet, pod *corev1.Pod) bool {
	parent, ordinal := getParentNameAndOrdinal(pod)
	return ordinal >= 0 &&
		set.Name == parent &&
		pod.Name == getPodName(set, ordinal) &&
		pod.Namespace == set.Namespace &&
		pod.Labels[apis.AdvancedStatefulSetPodNameLabel] == pod.Name
}

// storageMatches checks if a Pod's storage matches a StatefulSet.
func storageMatches(set *apis.AdvancedStatefulSet, pod *corev1.Pod) bool {
	ordinal := getOrdinal(pod)
	if ordinal < 0 {
		return false
	}
	volumes := make(map[string]corev1.Volume, len(pod.Spec.Volumes))
	for _, volume := range pod.Spec.Volumes {
		volumes[volume.Name] = volume
	}
	for _, claim := range set.Spec.VolumeClaimTemplates {
		volume, found := volumes[claim.Name]
		if !found ||
			volume.VolumeSource.PersistentVolumeClaim == nil ||
			volume.VolumeSource.PersistentVolumeClaim.ClaimName != getPersistentVolumeClaimName(set, &claim, ordinal) {
			return false
		}
	}
	return true
}

// getPersistentVolumeClaims gets the PersistentVolumeClaims for a Pod.
func getPersistentVolumeClaims(set *apis.AdvancedStatefulSet, pod *corev1.Pod) map[string]corev1.PersistentVolumeClaim {
	ordinal := getOrdinal(pod)
	templates := set.Spec.VolumeClaimTemplates
	claims := make(map[string]corev1.PersistentVolumeClaim, len(templates))
	for i := range templates {
		claim := templates[i]
		claim.Name = getPersistentVolumeClaimName(set, &claim, ordinal)
		claim.Namespace = set.Namespace
		if claim.Labels != nil {
			for key, value := range set.Spec.Selector.MatchLabels {
				claim.Labels[key] = value
			}
		} else {
			claim.Labels = set.Spec.Selector.MatchLabels
		}
		claims[templates[i].Name] = claim
	}
	return claims
}

// updateStorage updates a Pod's storage to match a StatefulSet.
func updateStorage(set *apis.AdvancedStatefulSet, pod *corev1.Pod) {
	currentVolumes := pod.Spec.Volumes
	claims := getPersistentVolumeClaims(set, pod)
	newVolumes := make([]corev1.Volume, 0, len(claims))
	for name, claim := range claims {
		newVolumes = append(newVolumes, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
					ReadOnly:  false,
				},
			},
		})
	}
	for i := range currentVolumes {
		if _, ok := claims[currentVolumes[i].Name]; !ok {
			newVolumes = append(newVolumes, currentVolumes[i])
		}
	}
	pod.Spec.Volumes = newVolumes
}

// initIdentity initializes a Pod's identity for a StatefulSet.
func initIdentity(set *apis.AdvancedStatefulSet, pod *corev1.Pod) {
	updateIdentity(set, pod)
	// Set these immutable fields only on initial Pod creation, not updates.
	pod.Spec.Hostname = pod.Name
	pod.Spec.Subdomain = set.Spec.ServiceName
}

// updateIdentity updates a Pod's identity to match a StatefulSet.
func updateIdentity(set *apis.AdvancedStatefulSet, pod *corev1.Pod) {
	pod.Name = getPodName(set, getOrdinal(pod))
	pod.Namespace = set.Namespace
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[apis.AdvancedStatefulSetPodNameLabel] = pod.Name
}

// isRunningAndReady checks if a Pod is running and ready.
func isRunningAndReady(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning && IsPodReady(pod)
}

// newStatefulSetPod creates a new Pod for a StatefulSet.
func newStatefulSetPod(set *apis.AdvancedStatefulSet, ordinal int) *corev1.Pod {
	var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v2"}
	controllerKind := SchemeGroupVersion.WithKind("AdvancedStatefulset")
	pod, _ := GetPodFromTemplate(&set.Spec.Template, set, metav1.NewControllerRef(set, controllerKind))
	pod.Name = getPodName(set, ordinal)
	initIdentity(set, pod)
	updateStorage(set, pod)
	return pod
}

// newVersionedStatefulSetPod creates a new Pod for a StatefulSet with a specific revision.
func newVersionedStatefulSetPod(currentSet, updateSet *apis.AdvancedStatefulSet, currentRevision, updateRevision string, ordinal int) *corev1.Pod {
	if currentSet.Spec.UpdateStrategy.Type == apis.RollingUpdateStatefulSetStrategyType &&
		(currentSet.Spec.UpdateStrategy.RollingUpdate == nil && ordinal < int(currentSet.Status.CurrentReplicas)) ||
		(currentSet.Spec.UpdateStrategy.RollingUpdate != nil && ordinal < int(*currentSet.Spec.UpdateStrategy.RollingUpdate.Partition)) {
		pod := newStatefulSetPod(currentSet, ordinal)
		setPodRevision(pod, currentRevision)
		return pod
	}
	pod := newStatefulSetPod(updateSet, ordinal)
	setPodRevision(pod, updateRevision)
	return pod
}

// ascendingOrdinal sorts Pods by ordinal.
type ascendingOrdinal []*corev1.Pod

func (ao ascendingOrdinal) Len() int {
	return len(ao)
}

func (ao ascendingOrdinal) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}

func (ao ascendingOrdinal) Less(i, j int) bool {
	return getOrdinal(ao[i]) < getOrdinal(ao[j])
}

// ShrinkPod scales down specific pods in the StatefulSet.

func (r *StatefulSetReconciler) PatchPodForScaleIn(set *apis.AdvancedStatefulSet, replicas int) (int, error) {
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	curNum := set.Status.CurrentReplicas
	needToScaleInNum := int(curNum) - replicas
	pods, err := r.getPodsForStatefulSet(set, selector)
	if err != nil {
		return -1, err
	}

	if needToScaleInNum <= 0 {
		return -1, nil
	}

	// Shuffle the pods to select randomly.
	shuffledPods := make([]*corev1.Pod, len(pods))
	perm := rand.Perm(len(pods))
	for i, v := range perm {
		shuffledPods[v] = pods[i]
	}

	selectedPods := shuffledPods[:needToScaleInNum]
	for _, pod := range selectedPods {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations[cacheworkerset.NeedScaleInAnnoKey] = "true"
		err := r.client.Update(context.TODO(), pod)
		if err != nil {
			return -1, err
		}
	}

	return needToScaleInNum, nil
}

func (r *StatefulSetReconciler) ScaleInPodFunc(set *apis.AdvancedStatefulSet, NeedReplicas int) (int, error) {

	PodScaleInNum, err := r.PatchPodForScaleIn(set, NeedReplicas)
	if PodScaleInNum == -1 || err != nil {
		return -1, err
	}

	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	pods, err := r.getPodsForStatefulSet(set, selector)

	if err != nil {
		return -1, err
	}

	// Initialize an empty slice to hold the ordinals of pods that need to be scaled in.
	podOrdinals := []int{}

	for _, pod := range pods {
		if pod.Annotations != nil && pod.Annotations[cacheworkerset.NeedScaleInAnnoKey] == "true" {
			podOrdinal := getOrdinal(pod)
			podOrdinals = append(podOrdinals, podOrdinal)
		}
	}
	targetPods := findPodByOrdinalAndAnnotation(pods, podOrdinals, string(cacheworkerset.WorkerTypeAnnoKey))
	if len(targetPods) == 0 {
		klog.Errorf("Pods with needToScaleIn annotation not found for StatefulSet %s/%s", set.Namespace, set.Name)
		return -1, nil
	}

	for _, targetPod := range targetPods {
		if isTerminating(targetPod) {
			klog.Infof("Pod %s for StatefulSet %s/%s is already terminating", targetPod.Name, set.Namespace, set.Name)
			continue
		}

		err = r.PodControl.DeleteStatefulPod(set, targetPod)
		if err != nil {
			klog.Errorf("Failed to delete pod %s for StatefulSet %s/%s: %v", targetPod.Name, set.Namespace, set.Name, err)
			return -1, err
		}

		klog.Infof("Successfully initiated termination of pod %s for StatefulSet %s/%s", targetPod.Name, set.Namespace, set.Name)
	}

	return PodScaleInNum, nil
}

// findPodByOrdinalAndAnnotation finds pods by their ordinal and annotation in a list of pods.

func findPodByOrdinalAndAnnotation(pods []*corev1.Pod, ordinals []int, annotationKey string) []*corev1.Pod {
	result := []*corev1.Pod{}
	for _, pod := range pods {
		podOrdinal := getOrdinal(pod)
		for _, ordinal := range ordinals {
			if podOrdinal == ordinal && pod.Annotations[annotationKey] == string(cacheworkerset.AdvancedStatefulSetType) {
				result = append(result, pod)
			}
		}
	}
	return result
}
