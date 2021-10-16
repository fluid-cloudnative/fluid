package kubeclient

import (
	"context"
	"regexp"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// controllerKind           = appsv1alpha1.SchemeGroupVersion.WithKind("BroadcastJob")
	statefulSetControllerKind = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
)

// Get pods of the specified statefulset
func GetPodsForStatefulSet(c client.Client, set *appsv1.StatefulSet, selector labels.Selector) (pods []*corev1.Pod, err error) {

	podList := &corev1.PodList{}

	err = c.List(context.TODO(), podList, &client.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		log.Error(err, "Failed to list pods for statefulset")
		return
	}

	for _, pod := range podList.Items {
		if isMemberOf(set, &pod) {
			controllerRef := metav1.GetControllerOf(&pod)
			if controllerRef != nil {
				// No controller should care about orphans being deleted.

				resolveControllerRef(c, controllerRef, set.Namespace, statefulSetControllerKind)
			} else {

			}
		} else {

		}
	}

	return
}

// statefulPodRegex is a regular expression that extracts the parent StatefulSet and ordinal from the Name of a Pod
var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the name of pod's parent StatefulSet and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSet, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
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

// getParentName gets the name of pod's parent StatefulSet. If pod has not parent, the empty string is returned.
func getParentName(pod *v1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

// isMemberOf tests if pod is a member of set.
func isMemberOf(set *appsv1.StatefulSet, pod *v1.Pod) bool {
	return getParentName(pod) == set.Name
}
