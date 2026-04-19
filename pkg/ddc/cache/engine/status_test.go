/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	testStatusNamespace  = "default"
	testStatusRuntime    = "curvine-demo"
	testStatusMaster     = "curvine-demo-master"
	testStatusWorker     = "curvine-demo-worker"
	testStatusClient     = "curvine-demo-client"
	testCacheRuntimeGR   = "cacheruntimes"
	testCacheRuntimeGV   = "data.fluid.io"
	testStatusWorkloadAP = "apps/v1"
)

func TestCheckAndUpdateRuntimeStatusClientNotReadyDoesNotBlockRuntimeReady(t *testing.T) {
	engine, client := newStatusTestEngineWithClient(
		t,
		fake.NewFakeClientWithScheme(
			datav1alpha1.UnitTestScheme,
			newStatusTestRuntime(),
			newStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
			newStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
			newDaemonSetComponent(testStatusClient, testStatusNamespace, 1, 0),
		),
	)

	ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to become ready once master and worker are ready")
	}

	updatedRuntime := getUpdatedRuntime(t, client)
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseNotReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseNotReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration == "" {
		t.Fatalf("expected setup duration to be recorded once runtime is ready")
	}
}

func TestCheckAndUpdateRuntimeStatusClientPartialReadyDoesNotBlockRuntimeReady(t *testing.T) {
	engine, client := newStatusTestEngineWithClient(
		t,
		fake.NewFakeClientWithScheme(
			datav1alpha1.UnitTestScheme,
			newStatusTestRuntime(),
			newStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
			newStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
			newDaemonSetComponent(testStatusClient, testStatusNamespace, 2, 1),
		),
	)

	ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to become ready once master and worker are ready")
	}

	updatedRuntime := getUpdatedRuntime(t, client)
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhasePartialReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhasePartialReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration == "" {
		t.Fatalf("expected setup duration to be recorded once runtime is ready")
	}
}

func TestCheckAndUpdateRuntimeStatusClientZeroDesiredReplicasReportsReady(t *testing.T) {
	engine, client := newStatusTestEngineWithClient(
		t,
		fake.NewFakeClientWithScheme(
			datav1alpha1.UnitTestScheme,
			newStatusTestRuntime(),
			newStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
			newStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
			newDaemonSetComponent(testStatusClient, testStatusNamespace, 0, 0),
		),
	)

	ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to stay ready when client desires zero replicas")
	}

	updatedRuntime := getUpdatedRuntime(t, client)
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.Client.DesiredReplicas != 0 {
		t.Fatalf("expected desired replicas to stay 0, got %d", updatedRuntime.Status.Client.DesiredReplicas)
	}
}

func TestCheckAndUpdateRuntimeStatusClientFullyReadyReportsReady(t *testing.T) {
	engine, client := newStatusTestEngineWithClient(
		t,
		fake.NewFakeClientWithScheme(
			datav1alpha1.UnitTestScheme,
			newStatusTestRuntime(),
			newStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
			newStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
			newDaemonSetComponent(testStatusClient, testStatusNamespace, 2, 2),
		),
	)

	ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to stay ready when client is fully ready")
	}

	updatedRuntime := getUpdatedRuntime(t, client)
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.Client.ReadyReplicas != updatedRuntime.Status.Client.DesiredReplicas {
		t.Fatalf("expected ready replicas to match desired replicas, got %d/%d", updatedRuntime.Status.Client.ReadyReplicas, updatedRuntime.Status.Client.DesiredReplicas)
	}
}

func TestCheckAndUpdateRuntimeStatusRecomputesRuntimeReadyOnRetry(t *testing.T) {
	baseClient := fake.NewFakeClientWithScheme(
		datav1alpha1.UnitTestScheme,
		newStatusTestRuntime(),
		newStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
		newStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
	)

	client := &conflictOnceClient{
		Client: baseClient,
		statusWriter: &conflictOnceStatusWriter{
			StatusWriter: baseClient.Status(),
			beforeConflict: func(ctx context.Context) error {
				worker := &appsv1.StatefulSet{}
				if err := baseClient.Get(ctx, types.NamespacedName{Name: testStatusWorker, Namespace: testStatusNamespace}, worker); err != nil {
					return err
				}

				worker.Status.ReadyReplicas = 0
				worker.Status.AvailableReplicas = 0
				return baseClient.Status().Update(ctx, worker)
			},
		},
	}

	engine, _ := newStatusTestEngineWithClient(t, client)
	ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(false))
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if ready {
		t.Fatalf("expected runtime to be not ready after retry sees worker become not ready")
	}

	updatedRuntime := getUpdatedRuntime(t, client)
	if updatedRuntime.Status.Worker.Phase != datav1alpha1.RuntimePhaseNotReady {
		t.Fatalf("expected worker phase %q, got %q", datav1alpha1.RuntimePhaseNotReady, updatedRuntime.Status.Worker.Phase)
	}
	if updatedRuntime.Status.SetupDuration != "" {
		t.Fatalf("expected setup duration to stay empty when final runtime status is not ready, got %q", updatedRuntime.Status.SetupDuration)
	}
}

func newStatusTestEngineWithClient(t *testing.T, client ctrlclient.Client) (*CacheEngine, ctrlclient.Client) {
	t.Helper()

	return &CacheEngine{
		Client:    client,
		Scheme:    datav1alpha1.UnitTestScheme,
		name:      testStatusRuntime,
		namespace: testStatusNamespace,
		Log:       fake.NullLogger(),
	}, client
}

func newStatusTestRuntimeValue(enableClient bool) *common.CacheRuntimeValue {
	value := &common.CacheRuntimeValue{
		Master: newStatusTestComponentValue(testStatusMaster, "StatefulSet"),
		Worker: newStatusTestComponentValue(testStatusWorker, "StatefulSet"),
		Client: newStatusTestComponentValue(testStatusClient, "DaemonSet"),
	}
	value.Client.Enabled = enableClient

	return value
}

func newStatusTestComponentValue(name, kind string) *common.CacheRuntimeComponentValue {
	return &common.CacheRuntimeComponentValue{
		Enabled:      true,
		Name:         name,
		Namespace:    testStatusNamespace,
		WorkloadType: metav1.TypeMeta{APIVersion: testStatusWorkloadAP, Kind: kind},
	}
}

func newStatusTestRuntime() *datav1alpha1.CacheRuntime {
	return &datav1alpha1.CacheRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:              testStatusRuntime,
			Namespace:         testStatusNamespace,
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Minute)),
		},
	}
}

func getUpdatedRuntime(t *testing.T, client ctrlclient.Client) *datav1alpha1.CacheRuntime {
	t.Helper()

	updatedRuntime := &datav1alpha1.CacheRuntime{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: testStatusRuntime, Namespace: testStatusNamespace}, updatedRuntime); err != nil {
		t.Fatalf("failed to get updated runtime: %v", err)
	}

	return updatedRuntime
}

func newStatefulSetComponent(name, namespace string, desiredReplicas, readyReplicas int32) *appsv1.StatefulSet {
	replicas := desiredReplicas
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			CurrentReplicas:   desiredReplicas,
			AvailableReplicas: readyReplicas,
			ReadyReplicas:     readyReplicas,
		},
	}
}

func newDaemonSetComponent(name, namespace string, desiredReplicas, readyReplicas int32) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: appsv1.DaemonSetStatus{
			CurrentNumberScheduled: desiredReplicas,
			DesiredNumberScheduled: desiredReplicas,
			NumberAvailable:        readyReplicas,
			NumberReady:            readyReplicas,
			NumberUnavailable:      desiredReplicas - readyReplicas,
		},
	}
}

type conflictOnceClient struct {
	ctrlclient.Client
	statusWriter ctrlclient.StatusWriter
}

func (c *conflictOnceClient) Status() ctrlclient.StatusWriter {
	return c.statusWriter
}

type conflictOnceStatusWriter struct {
	ctrlclient.StatusWriter
	beforeConflict func(ctx context.Context) error
	conflicted     bool
}

func (w *conflictOnceStatusWriter) Update(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.SubResourceUpdateOption) error {
	if !w.conflicted {
		w.conflicted = true
		if w.beforeConflict != nil {
			if err := w.beforeConflict(ctx); err != nil {
				return err
			}
		}

		return apierrors.NewConflict(
			schema.GroupResource{Group: testCacheRuntimeGV, Resource: testCacheRuntimeGR},
			obj.GetName(),
			errors.New("injected conflict"),
		)
	}

	return w.StatusWriter.Update(ctx, obj, opts...)
}
