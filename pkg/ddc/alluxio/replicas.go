/*
Copyright 2023 The Fluid Authors.

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

package alluxio

import (
	"context"
	"fmt"
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	openkruise "github.com/openkruise/kruise-api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ScaleStatefulSet scale the statefulset replicas
func ScaleAdvancedStatefulSet(client client.Client, name string, namespace string, replicas int32) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := GetAdvancedStatefulSet(client, name, namespace)
		if err != nil {
			return err
		}
		workersToUpdate := workers.DeepCopy()
		workersToUpdate.Spec.Replicas = &replicas
		if !reflect.DeepEqual(workers, workersToUpdate) {
			err = client.Update(context.TODO(), workersToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetStatefulset gets the statefulset by name and namespace
func GetAdvancedStatefulSet(c client.Client, name string, namespace string) (master *openkruise.StatefulSet, err error) {
	master = &openkruise.StatefulSet{}
	err = c.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}

func GetWorkersAsAdvancedStatefulset(client client.Client, key types.NamespacedName) (workers *openkruise.StatefulSet, err error) {
	workers, err = GetAdvancedStatefulSet(client, key.Name, key.Namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			_, dsErr := kubeclient.GetDaemonset(client, key.Name, key.Namespace)
			// return workers, fluiderr.NewDeprecated()
			// find the daemonset successfully
			if dsErr == nil {
				return workers, fluiderrs.NewDeprecated(schema.GroupResource{
					Group:    appsv1.SchemeGroupVersion.Group,
					Resource: "daemonsets",
				}, key)
			}
		}
	}

	return
}

// SyncReplicas syncs the replicas
func (e *AlluxioEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
			if fluiderrs.IsDeprecated(err) {
				e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, scale out/in are not supported. To support these features, please create a new dataset", "details", err)
				return nil
			}
			if errors.IsNotFound(err) {
				cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
					fmt.Sprintf("The statefulset %s in %s is not found, please fix it.",
						e.getWorkerName(),
						e.namespace), corev1.ConditionFalse)

				updateErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {

					runtime, err := e.getRuntime()
					if err != nil {
						return err
					}

					runtimeToUpdate := runtime.DeepCopy()

					_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

					if oldCond == nil || oldCond.Type != cond.Type {
						runtimeToUpdate.Status.Conditions =
							utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
								cond)
					}

					runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
					e.Log.Error(err, "the worker are not ready")

					if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
						updateErr := e.Client.Status().Update(context.TODO(), runtimeToUpdate)
						if updateErr != nil {
							return updateErr
						}

						updateErr = e.UpdateDatasetStatus(data.FailedDatasetPhase)
						if updateErr != nil {
							e.Log.Error(updateErr, "Failed to update dataset")
							return updateErr
						}
					}

					return err
				})
				totalErr := fmt.Errorf("the master engine is not existed %v", updateErr)
				return totalErr
			}
			return err
		}
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		if runtime.Spec.ScaleConfig.WorkerType == cacheworkerset.AdvancedStatefulSetType {
			logf.Log.Info("SyncReplicas for AdvancedStatefulSet")
			workers, err := GetWorkersAsAdvancedStatefulset(e.Client,
				types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
			if err != nil {
				return err
			}

			err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)

		} else {
			err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		}

		return err
	})
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to sync replicas", types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}
