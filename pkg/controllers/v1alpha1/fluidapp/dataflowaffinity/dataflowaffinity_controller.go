/*
 Copyright 2024 The Fluid Authors.

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

package dataflowaffinity

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

const DataOpJobControllerName string = "DataOpJobController"

type DataOpJobReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
}

func (f *DataOpJobReconciler) ControllerName() string {
	return DataOpJobControllerName
}

func (f *DataOpJobReconciler) ManagedResource() client.Object {
	return &batchv1.Job{}
}

type reconcileRequestContext struct {
	context.Context
	Log logr.Logger
	job *batchv1.Job
	types.NamespacedName
}

func NewDataOpJobReconciler(client client.Client,
	log logr.Logger,
	recorder record.EventRecorder) *DataOpJobReconciler {
	return &DataOpJobReconciler{
		Client:   client,
		Recorder: recorder,
		Log:      log,
	}
}

// Reconcile reconciles Jobs
// +kubebuilder:rbac:groups=v1,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=pods/status,verbs=get;update;patch
func (f *DataOpJobReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	requestCtx := reconcileRequestContext{
		Context:        ctx,
		Log:            f.Log.WithValues("namespacedName", request.NamespacedName),
		NamespacedName: request.NamespacedName,
	}
	job, err := kubeclient.GetJob(f.Client, request.Name, request.Namespace)
	if err != nil {
		requestCtx.Log.Error(err, "fetch job error")
		return reconcile.Result{}, err
	}
	if job == nil {
		requestCtx.Log.Info("job not found", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}
	requestCtx.job = job

	if !watch.JobShouldInQueue(job) {
		requestCtx.Log.Info("job should not in queue", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}

	// inject dataflow enabled affinity if not exist.
	if _, ok := job.Annotations[common.AnnotationDataFlowAffinityInject]; !ok {
		job.Annotations[common.AnnotationDataFlowAffinityInject] = "true"
		if err := f.Client.Update(ctx, job); err != nil {
			requestCtx.Log.Error(err, "Failed to add dataflow affinity enabled label", "AnnotationUpdateError", ctx)
			return utils.RequeueIfError(err)
		}
	}
	// get job' status, if succeed, add label to job.
	condition := kubeclient.GetFinishedJobCondition(job)
	if condition != nil && condition.Type == batchv1.JobComplete {
		err = f.injectPodNodeLabelsToJob(job)
		if err != nil {
			requestCtx.Log.Error(err, "update labels for job failed")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (f *DataOpJobReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return watch.SetupDataOpJobWatcherWithReconciler(mgr, options, f)
}

func (f *DataOpJobReconciler) injectPodNodeLabelsToJob(job *batchv1.Job) error {
	pod, err := kubeclient.GetSucceedPodForJob(f.Client, job)
	if err != nil {
		return err
	}
	if pod == nil {
		return fmt.Errorf("completed job has no succeed pod, jobNamespace: %s, jobName: %s", job.Namespace, job.Name)
	}

	nodeName := pod.Spec.NodeName
	if len(nodeName) == 0 {
		return fmt.Errorf("succeed job has no node name, podNamespace: %s, podName: %s", pod.Namespace, pod.Name)
	}

	node, err := kubeclient.GetNode(f.Client, nodeName)
	if err != nil {
		return fmt.Errorf("error to get node %s: %v", nodeName, err)
	}

	injectLabels := map[string]string{}
	// node
	injectLabels[common.K8sNodeNameLabelKey] = nodeName
	// region
	region, exist := node.Labels[common.K8sRegionLabelKey]
	if exist {
		injectLabels[common.K8sRegionLabelKey] = region
	}
	// zone
	zone, exist := node.Labels[common.K8sZoneLabelKey]
	if exist {
		injectLabels[common.K8sZoneLabelKey] = zone
	}

	// customized labels
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		fillCustomizedNodeAffinity(pod.Spec.Affinity.NodeAffinity, injectLabels, node)
	}

	// update job labels, reconciled job is selected by labels so the field will not be nil.
	for k, v := range injectLabels {
		job.Labels[k] = v
	}
	if err = f.Client.Update(context.TODO(), job); err != nil {
		return err
	}

	return nil
}

func fillCustomizedNodeAffinity(podNodeAffinity *corev1.NodeAffinity, injectLabels map[string]string, node *corev1.Node) {
	// prefer
	for _, term := range podNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		for _, expression := range term.Preference.MatchExpressions {
			// use the actually value in the node. Transform In, NotIn, Exists, DoesNotExist. Gt, and Lt to In.
			value, exist := node.Labels[expression.Key]
			if exist {
				// add customized prefix to distinguish
				injectLabels[common.LabelDataFlowAffinityPrefix+expression.Key] = value
			}
		}
	}

	if podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return
	}

	// require
	for _, term := range podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, expression := range term.MatchExpressions {
			// use the actually value in the node. Transform In, NotIn, Exists, DoesNotExist. Gt, and Lt to In.
			value, exist := node.Labels[expression.Key]
			if exist {
				// add customized prefix to distinguish
				injectLabels[common.LabelDataFlowAffinityPrefix+expression.Key] = value
			}
		}
	}
}
