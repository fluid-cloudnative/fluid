package apis

import (
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"context"
	asv1 "github.com/pingcap/advanced-statefulset/client/apis/apps/v1"
	asclientset "github.com/pingcap/advanced-statefulset/client/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// UpgradeToAdvancedStatefulSetAnn represents the annotation key used to
	// help migration to Advanced StatefulSet
	UpgradeToAdvancedStatefulSetAnn = "apps.pingcap.com/upgrade-to-asts"
)

// Upgrade upgrades Kubernetes builtin StatefulSet to Advanced StatefulSet.
//
// This method is idempotent. The last operation is deleting the builtin
// StatefulSet, the caller must retry until the builtin StatefulSet is deleted
// succesfully.
//
// Basic procedure:
//
// - remove sts selector lebels from controller revisions and set a special annotation for Advanced StatefulSet (can be skipped if Kubernetes cluster has http://issues.k8s.io/84982 fixed)
// - create advanced sts
// - delete sts with DeletePropagationOrphan policy
func Upgrade(ctx context.Context, c clientset.Interface, asc asclientset.Interface, sts *appsv1.StatefulSet) (*asv1.StatefulSet, error) {
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, err
	}
	// It's important to empty statefulset selector labels,
	// otherwise sts will adopt it again on delete event and then
	// GC will delete revisions because they are not orphans.
	// https://github.com/kubernetes/kubernetes/issues/84982
	revisionListOptions := metav1.ListOptions{LabelSelector: selector.String()}
	oldRevisionList, err := c.AppsV1().ControllerRevisions(sts.Namespace).List(ctx, revisionListOptions)
	if err != nil {
		return nil, err
	}
	for _, revision := range oldRevisionList.Items {
		for key := range sts.Spec.Selector.MatchLabels {
			delete(revision.Labels, key)
		}
		revision.Labels[UpgradeToAdvancedStatefulSetAnn] = sts.Name
		_, err = c.AppsV1().ControllerRevisions(revision.Namespace).Update(ctx, &revision, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}
	klog.V(2).Infof("Succesfully marked all controller revisions (%d) of StatefulSet %s/%s", len(oldRevisionList.Items), sts.Namespace, sts.Name)

	// Create or Update
	asts, err := asc.AppsV1().StatefulSets(sts.Namespace).Get(ctx, sts.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	notFound := apierrors.IsNotFound(err)
	upgradedSts, err := FromBuiltinStatefulSet(sts)
	if err != nil {
		return nil, err
	}
	if notFound {
		asts = upgradedSts.DeepCopy()
		// https://github.com/kubernetes/apiserver/blob/kubernetes-1.16.0/pkg/storage/etcd3/store.go#L141-L143
		asts.ObjectMeta.ResourceVersion = ""
		// https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply
		// old ManagedFields belongs to apps/v1 and kube-controller-manager,
		// nil it and the ownership will be transferred to
		// advanced-statefulset-controller-manager
		asts.ObjectMeta.ManagedFields = nil
		asts, err = asc.AppsV1().StatefulSets(asts.Namespace).Create(ctx, asts, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		klog.V(2).Infof("Succesfully created the new Advanced StatefulSet %s/%s", asts.Namespace, asts.Name)
	} else {
		asts.Spec = upgradedSts.Spec
		asts, err = asc.AppsV1().StatefulSets(asts.Namespace).Update(ctx, asts, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		klog.V(2).Infof("Succesfully updated the Advanced StatefulSet %s/%s", asts.Namespace, asts.Name)
	}

	// Status must be updated via UpdateStatus
	asts.Status = upgradedSts.Status
	asts, err = asc.AppsV1().StatefulSets(asts.Namespace).UpdateStatus(ctx, asts, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	// At the last, delete the builtin StatefulSet
	policy := metav1.DeletePropagationOrphan
	err = c.AppsV1().StatefulSets(sts.Namespace).Delete(ctx, sts.Name, metav1.DeleteOptions{
		PropagationPolicy: &policy,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		// ignore IsNotFound error
		return nil, err
	}
	klog.V(2).Infof("Succesfully deleted the old builtin StatefulSet %s/%s", sts.Namespace, sts.Name)
	return asts, nil
}
