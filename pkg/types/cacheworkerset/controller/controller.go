package statefulset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	helper "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/apis"
	apps "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/client/v1"
	"github.com/go-logr/logr"
	"hash/fnv"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
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
)

// StatefulSetReconciler reconciles a StatefulSet object.
type StatefulSetReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// Reconcile is the main reconciliation loop for StatefulSet.
func (r *StatefulSetReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	//namespace, name, err := cache.SplitMetaNamespaceKey(req.NamespacedName.String())
	//if err != nil {
	//	return reconcile.Result{}, err
	//}
	set := &apps.AdvancedStatefulSet{}
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
func (r *StatefulSetReconciler) ListRevisions(set *apps.AdvancedStatefulSet) ([]*appsv1.ControllerRevision, error) {
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		return nil, err
	}
	revisions := &appsv1.ControllerRevisionList{}
	err = r.Client.List(context.TODO(), revisions, client.InNamespace(set.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	revisinsToUpgrade, err := r.Client.List(context.TODO(), revisions, client.InNamespace(set.Namespace), client.MatchingLabels(map[string]string{helper.UpgradeToAdvancedStatefulSetAnn: set.Name}))
	if err != nil {
		return nil, err
	}
	res := []*appsv1.ControllerRevision{}
	for _, item := range append(revisions.Items, revisinsToUpgrade.Items...) {
		local := item
		res = append(res, &local)
	}
	return res, nil
}

// getStatefulSetRevisions returns the current and update ControllerRevisions for set.
func (r *StatefulSetReconciler) getStatefulSetRevisions(set *apps.AdvancedStatefulSet, revisions []*appsv1.ControllerRevision) (*appsv1.ControllerRevision, *appsv1.ControllerRevision, int32, error) {
	var currentRevision, updateRevision *appsv1.ControllerRevision

	revisionCount := len(revisions)
	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Revision < revisions[j].Revision
	})

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
func (r *StatefulSetReconciler) updateStatefulSet(set *apps.AdvancedStatefulSet, currentRevision *appsv1.ControllerRevision, updateRevision *appsv1.ControllerRevision, collisionCount int32, pods []*corev1.Pod) (*apps.AdvancedStatefulSetStatus, error) {
	currentSet, err := applyRevision(set, currentRevision)
	if err != nil {
		return nil, err
	}
	updateSet, err := applyRevision(set, updateRevision)
	if err != nil {
		return nil, err
	}

	status := apps.AdvancedStatefulSetStatus{}
	status.ObservedGeneration = set.Generation
	status.CurrentRevision = currentRevision.Name
	status.UpdateRevision = updateRevision.Name
	status.CollisionCount = new(int32)
	*status.CollisionCount = collisionCount

	deleteSlots := helper.GetDeleteSlots(set)
	_replicaCount, deleteSlots := helper.GetMaxReplicaCountAndDeleteSlots(*set.Spec.Replicas, deleteSlots)
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
			if err := r.podControl.DeleteStatefulPod(set, replicas[i]); err != nil {
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
			if err := r.podControl.CreateStatefulPod(set, replicas[i]); err != nil {
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
		if err := r.podControl.UpdateStatefulPod(updateSet, replica); err != nil {
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

		if err := r.podControl.DeleteStatefulPod(set, condemned[target]); err != nil {
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

	if set.Spec.UpdateStrategy.Type == apps.OnDeleteStatefulSetStrategyType {
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
			err := r.podControl.DeleteStatefulPod(set, replicas[target])
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
func (r *StatefulSetReconciler) updateStatefulSetStatus(set *apps.AdvancedStatefulSet, status *apps.AdvancedStatefulSetStatus) error {
	completeRollingUpdate(set, status)

	if !inconsistentStatus(set, status) {
		return nil
	}

	set = set.DeepCopy()
	return r.Client.Status().Update(context.TODO(), set)
}

// podControl is used for patching pods.
type podControl struct {
	Client   client.Client
	Recorder record.EventRecorder
}

var _ StatefulPodControlInterface = &podControl{}

// CreateStatefulPod creates a Pod in a StatefulSet.
func (pc *podControl) CreateStatefulPod(set *apps.AdvancedStatefulSet, pod *corev1.Pod) error {
	err := pc.Client.Create(context.TODO(), pod)
	if err != nil {
		pc.Recorder.Eventf(set, corev1.EventTypeWarning, "FailedToCreatePod", "Failed to create Pod %s for StatefulSet %s/%s: %v", pod.Name, set.Namespace, set.Name, err)
		return err
	}
	pc.Recorder.Eventf(set, corev1.EventTypeNormal, "CreatedPod", "Created Pod %s for StatefulSet %s/%s", pod.Name, set.Namespace, set.Name)
	return nil
}

// DeleteStatefulPod deletes a Pod in a StatefulSet.
func (pc *podControl) DeleteStatefulPod(set *apps.AdvancedStatefulSet, pod *corev1.Pod) error {
	err := pc.Client.Delete(context.TODO(), pod)
	if err != nil {
		pc.Recorder.Eventf(set, corev1.EventTypeWarning, "FailedToDeletePod", "Failed to delete Pod %s for StatefulSet %s/%s: %v", pod.Name, set.Namespace, set.Name, err)
		return err
	}
	pc.Recorder.Eventf(set, corev1.EventTypeNormal, "DeletedPod", "Deleted Pod %s for StatefulSet %s/%s", pod.Name, set.Namespace, set.Name)
	return nil
}

// UpdateStatefulPod updates a Pod in a StatefulSet.
func (pc *podControl) UpdateStatefulPod(set *apps.AdvancedStatefulSet, pod *corev1.Pod) error {
	err := pc.Client.Update(context.TODO(), pod)
	if err != nil {
		pc.Recorder.Eventf(set, corev1.EventTypeWarning, "FailedToUpdatePod", "Failed to update Pod %s for StatefulSet %s/%s: %v", pod.Name, set.Namespace, set.Name, err)
		return err
	}
	pc.Recorder.Eventf(set, corev1.EventTypeNormal, "UpdatedPod", "Updated Pod %s for StatefulSet %s/%s", pod.Name, set.Namespace, set.Name)
	return nil
}

// getPodsForStatefulSet returns the pods associated with a StatefulSet.
func (r *StatefulSetReconciler) getPodsForStatefulSet(set *apps.AdvancedStatefulSet, selector labels.Selector) ([]*corev1.Pod, error) {
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

// newRevision creates a new ControllerRevision for a StatefulSet.
func newRevision(set *apps.AdvancedStatefulSet, revision *appsv1.ControllerRevision, collisionCount *int32) (*appsv1.ControllerRevision, error) {
	patch, err := getPatch(set)
	if err != nil {
		return nil, err
	}
	cr := &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name: controllerRevisionName(set.Name, hashControllerRevision(revision, collisionCount)),
			Labels: map[string]string{
				helper.UpgradeToAdvancedStatefulSetAnn: set.Name,
			},
			Annotations: set.Annotations,
		},
		Revision: revision,
		Data: corev1.RawExtension{
			Raw: patch,
		},
	}
	return cr, nil
}

// applyRevision applies a ControllerRevision to a StatefulSet.
func applyRevision(set *apps.AdvancedStatefulSet, revision *appsv1.ControllerRevision) (*apps.AdvancedStatefulSet, error) {
	clone := set.DeepCopy()
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(scheme.Codecs, clone)), revision.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	restoredSet := &apps.AdvancedStatefulSet{}
	err = json.Unmarshal(patched, restoredSet)
	if err != nil {
		return nil, err
	}
	return restoredSet, nil
}

// nextRevision finds the next valid revision number.
func nextRevision(revisions []*appsv1.ControllerRevision) int64 {
	if len(revisions) == 0 {
		return 1
	}
	return revisions[len(revisions)-1].Revision + 1
}

// inconsistentStatus checks if a StatefulSet's status is inconsistent.
func inconsistentStatus(set *apps.AdvancedStatefulSet, status *apps.AdvancedStatefulSetStatus) bool {
	return status.ObservedGeneration > set.Status.ObservedGeneration ||
		status.Replicas != set.Status.Replicas ||
		status.CurrentReplicas != set.Status.CurrentReplicas ||
		status.ReadyReplicas != set.Status.ReadyReplicas ||
		status.UpdatedReplicas != set.Status.UpdatedReplicas ||
		status.CurrentRevision != set.Status.CurrentRevision ||
		status.UpdateRevision != set.Status.UpdateRevision
}

// completeRollingUpdate completes a rolling update for a StatefulSet.
func completeRollingUpdate(set *apps.AdvancedStatefulSet, status *apps.AdvancedStatefulSetStatus) {
	if set.Spec.UpdateStrategy.Type == apps.RollingUpdateStatefulSetStrategyType &&
		status.UpdatedReplicas == status.Replicas &&
		status.ReadyReplicas == status.Replicas {
		status.CurrentReplicas = status.UpdatedReplicas
		status.CurrentRevision = status.UpdateRevision
	}
}

// overlappingStatefulSets sorts StatefulSets by creation timestamp and name.
type overlappingStatefulSets []*apps.AdvancedStatefulSet

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

// getOrdinal gets the ordinal from a Pod.
func getOrdinal(pod *corev1.Pod) int {
	_, ordinal := getParentNameAndOrdinal(pod)
	return ordinal
}

// getPodName gets the name of a Pod for a StatefulSet and ordinal.
func getPodName(set *apps.AdvancedStatefulSet, ordinal int) string {
	return fmt.Sprintf("%s-%d", set.Name, ordinal)
}

// getPersistentVolumeClaimName gets the name of a PersistentVolumeClaim for a Pod.
func getPersistentVolumeClaimName(set *apps.AdvancedStatefulSet, claim *corev1.PersistentVolumeClaim, ordinal int) string {
	return fmt.Sprintf("%s-%s-%d", claim.Name, set.Name, ordinal)
}

// isMemberOf checks if a Pod is a member of a StatefulSet.
func isMemberOf(set *apps.AdvancedStatefulSet, pod *corev1.Pod) bool {
	return getParentName(pod) == set.Name
}

// identityMatches checks if a Pod has a valid identity for a StatefulSet.
func identityMatches(set *apps.AdvancedStatefulSet, pod *corev1.Pod) bool {
	parent, ordinal := getParentNameAndOrdinal(pod)
	return ordinal >= 0 &&
		set.Name == parent &&
		pod.Name == getPodName(set, ordinal) &&
		pod.Namespace == set.Namespace &&
		pod.Labels[apps.AdvancedStatefulSetPodNameLabel] == pod.Name
}

// storageMatches checks if a Pod's storage matches a StatefulSet.
func storageMatches(set *apps.AdvancedStatefulSet, pod *corev1.Pod) bool {
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
func getPersistentVolumeClaims(set *apps.AdvancedStatefulSet, pod *corev1.Pod) map[string]corev1.PersistentVolumeClaim {
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
func updateStorage(set *apps.AdvancedStatefulSet, pod *corev1.Pod) {
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
func initIdentity(set *apps.AdvancedStatefulSet, pod *corev1.Pod) {
	updateIdentity(set, pod)
	// Set these immutable fields only on initial Pod creation, not updates.
	pod.Spec.Hostname = pod.Name
	pod.Spec.Subdomain = set.Spec.ServiceName
}

// updateIdentity updates a Pod's identity to match a StatefulSet.
func updateIdentity(set *apps.AdvancedStatefulSet, pod *corev1.Pod) {
	pod.Name = getPodName(set, getOrdinal(pod))
	pod.Namespace = set.Namespace
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[apps.AdvancedStatefulSetPodNameLabel] = pod.Name
}

// isRunningAndReady checks if a Pod is running and ready.
func isRunningAndReady(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning && k8s.IsPodReady(pod)
}

// isCreated checks if a Pod has been created.
func isCreated(pod *corev1.Pod) bool {
	return pod.Status.Phase != ""
}

// isFailed checks if a Pod has failed.
func isFailed(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodFailed
}

// isSucceeded checks if a Pod has succeeded.
func isSucceeded(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded
}

// isTerminating checks if a Pod is being terminated.
func isTerminating(pod *corev1.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// isHealthy checks if a Pod is healthy.
func isHealthy(pod *corev1.Pod) bool {
	return isRunningAndReady(pod) && !isTerminating(pod)
}

// allowsBurst checks if a StatefulSet allows burst operations.
func allowsBurst(set *apps.AdvancedStatefulSet) bool {
	return set.Spec.PodManagementPolicy == apps.ParallelPodManagement
}

// setPodRevision sets the revision of a Pod.
func setPodRevision(pod *corev1.Pod, revision string) {
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[appsv1.StatefulSetRevisionLabel] = revision
}

// getPodRevision gets the revision of a Pod.
func getPodRevision(pod *corev1.Pod) string {
	if pod.Labels == nil {
		return ""
	}
	return pod.Labels[appsv1.StatefulSetRevisionLabel]
}

// newStatefulSetPod creates a new Pod for a StatefulSet.
func newStatefulSetPod(set *apps.AdvancedStatefulSet, ordinal int) *corev1.Pod {
	pod, _ := k8s.GetPodFromTemplate(&set.Spec.Template, set, metav1.NewControllerRef(set, controllerKind))
	pod.Name = getPodName(set, ordinal)
	initIdentity(set, pod)
	updateStorage(set, pod)
	return pod
}

// newVersionedStatefulSetPod creates a new Pod for a StatefulSet with a specific revision.
func newVersionedStatefulSetPod(currentSet, updateSet *apps.AdvancedStatefulSet, currentRevision, updateRevision string, ordinal int) *corev1.Pod {
	if currentSet.Spec.UpdateStrategy.Type == apps.RollingUpdateStatefulSetStrategyType &&
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

// hashControllerRevision hashes a ControllerRevision.
func hashControllerRevision(revision *appsv1.ControllerRevision, probe *int32) string {
	hf := fnv.New32()
	if len(revision.Data.Raw) > 0 {
		hf.Write(revision.Data.Raw)
	}
	if revision.Data.Object != nil {
		k8s.DeepHashObject(hf, revision.Data.Object)
	}
	if probe != nil {
		hf.Write([]byte(strconv.FormatInt(int64(*probe), 10)))
	}
	return rand.SafeEncodeString(fmt.Sprint(hf.Sum32()))
}

// controllerRevisionName generates a ControllerRevision name.
func controllerRevisionName(prefix string, hash string) string {
	if len(prefix) > 223 {
		prefix = prefix[:223]
	}
	return fmt.Sprintf("%s-%s", prefix, hash)
}

// getPatch generates a patch for a StatefulSet.
func getPatch(set *apps.AdvancedStatefulSet) ([]byte, error) {
	str, err := runtime.Encode(scheme.Codecs, set)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	json.Unmarshal([]byte(str), &raw)
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

func PerformController() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		panic(err)
	}

	recorder := mgr.GetEventRecorderFor("statefulset-controller")
	reconciler := &StatefulSetReconciler{
		Client:   mgr.GetClient(),
		Recorder: recorder,
	}

	err = reconciler.SetupWithManager(mgr)
	if err != nil {
		panic(err)
	}

	err = ctrl.NewWebhookManagedBy(mgr).For(&apps.AdvancedStatefulSet{}).Complete()
	if err != nil {
		panic(err)
	}

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		panic(err)
	}
}

// ShrinkPod scales down a specific pod in the StatefulSet.
func (r *StatefulSetReconciler) ShrinkPod(set *apps.AdvancedStatefulSet, podOrdinal int) error {
	pods, err := r.getPodsForStatefulSet(set, metav1.LabelSelectorAsSelector(set.Spec.Selector))
	if err != nil {
		return err
	}

	targetPod := findPodByOrdinal(pods, podOrdinal)
	if targetPod == nil {
		klog.Errorf("Pod with ordinal %d not found for StatefulSet %s/%s", podOrdinal, set.Namespace, set.Name)
		return nil
	}

	if isTerminating(targetPod) {
		klog.Infof("Pod %s for StatefulSet %s/%s is already terminating", targetPod.Name, set.Namespace, set.Name)
		return nil
	}

	err = r.podControl.DeleteStatefulPod(set, targetPod)
	if err != nil {
		klog.Errorf("Failed to delete pod %s for StatefulSet %s/%s: %v", targetPod.Name, set.Namespace, set.Name, err)
		return err
	}

	klog.Infof("Successfully initiated termination of pod %s for StatefulSet %s/%s", targetPod.Name, set.Namespace, set.Name)
	return nil
}

// findPodByOrdinal finds a pod by its ordinal in a list of pods.
func findPodByOrdinal(pods []*corev1.Pod, ordinal int) *corev1.Pod {
	for _, pod := range pods {
		if getOrdinal(pod) == ordinal {
			return pod
		}
	}
	return nil
}
