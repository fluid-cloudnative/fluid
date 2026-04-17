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
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestCheckAndUpdateRuntimeStatusRequiresClientReady(t *testing.T) {
	const (
		namespace   = "default"
		runtimeName = "curvine-demo"
		masterName  = "curvine-demo-master"
		workerName  = "curvine-demo-worker"
		clientName  = "curvine-demo-client"
	)

	runtime := &datav1alpha1.CacheRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:              runtimeName,
			Namespace:         namespace,
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Minute)),
		},
	}

	client := fake.NewFakeClientWithScheme(
		datav1alpha1.UnitTestScheme,
		runtime,
		newStatefulSetComponent(masterName, namespace, 1, 1),
		newStatefulSetComponent(workerName, namespace, 1, 1),
		newDaemonSetComponent(clientName, namespace, 1, 0),
	)

	engine := &CacheEngine{
		Client:    client,
		Scheme:    datav1alpha1.UnitTestScheme,
		name:      runtimeName,
		namespace: namespace,
		Log:       fake.NullLogger(),
	}

	value := &common.CacheRuntimeValue{
		Master: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         masterName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
		},
		Worker: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         workerName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
		},
		Client: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         clientName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "DaemonSet"},
		},
	}

	ready, err := engine.CheckAndUpdateRuntimeStatus(value)
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if ready {
		t.Fatalf("expected runtime to stay not ready while client is not ready")
	}

	updatedRuntime := &datav1alpha1.CacheRuntime{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: runtimeName, Namespace: namespace}, updatedRuntime); err != nil {
		t.Fatalf("failed to get runtime after first status update: %v", err)
	}
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseNotReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseNotReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration != "" {
		t.Fatalf("expected setup duration to stay empty before runtime is ready, got %q", updatedRuntime.Status.SetupDuration)
	}

	clientDaemonSet := &appsv1.DaemonSet{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: clientName, Namespace: namespace}, clientDaemonSet); err != nil {
		t.Fatalf("failed to get client daemonset: %v", err)
	}
	clientDaemonSet.Status.NumberReady = 1
	clientDaemonSet.Status.NumberAvailable = 1
	clientDaemonSet.Status.NumberUnavailable = 0
	if err := client.Status().Update(context.TODO(), clientDaemonSet); err != nil {
		t.Fatalf("failed to update client daemonset status: %v", err)
	}

	ready, err = engine.CheckAndUpdateRuntimeStatus(value)
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to become ready once client is ready")
	}

	if err := client.Get(context.TODO(), types.NamespacedName{Name: runtimeName, Namespace: namespace}, updatedRuntime); err != nil {
		t.Fatalf("failed to get runtime after second status update: %v", err)
	}
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration == "" {
		t.Fatalf("expected setup duration to be recorded once runtime is ready")
	}
}

func TestCheckAndUpdateRuntimeStatusRequiresAllClientReplicasReady(t *testing.T) {
	const (
		namespace   = "default"
		runtimeName = "curvine-demo"
		masterName  = "curvine-demo-master"
		workerName  = "curvine-demo-worker"
		clientName  = "curvine-demo-client"
	)

	runtime := &datav1alpha1.CacheRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:              runtimeName,
			Namespace:         namespace,
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Minute)),
		},
	}

	client := fake.NewFakeClientWithScheme(
		datav1alpha1.UnitTestScheme,
		runtime,
		newStatefulSetComponent(masterName, namespace, 1, 1),
		newStatefulSetComponent(workerName, namespace, 1, 1),
		newDaemonSetComponent(clientName, namespace, 2, 1),
	)

	engine := &CacheEngine{
		Client:    client,
		Scheme:    datav1alpha1.UnitTestScheme,
		name:      runtimeName,
		namespace: namespace,
		Log:       fake.NullLogger(),
	}

	value := &common.CacheRuntimeValue{
		Master: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         masterName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
		},
		Worker: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         workerName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
		},
		Client: &common.CacheRuntimeComponentValue{
			Enabled:      true,
			Name:         clientName,
			Namespace:    namespace,
			WorkloadType: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "DaemonSet"},
		},
	}

	ready, err := engine.CheckAndUpdateRuntimeStatus(value)
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if ready {
		t.Fatalf("expected runtime to stay not ready while client is only partially ready")
	}

	updatedRuntime := &datav1alpha1.CacheRuntime{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: runtimeName, Namespace: namespace}, updatedRuntime); err != nil {
		t.Fatalf("failed to get runtime after partial-ready status update: %v", err)
	}
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhasePartialReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhasePartialReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration != "" {
		t.Fatalf("expected setup duration to stay empty before all client replicas are ready, got %q", updatedRuntime.Status.SetupDuration)
	}

	clientDaemonSet := &appsv1.DaemonSet{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: clientName, Namespace: namespace}, clientDaemonSet); err != nil {
		t.Fatalf("failed to get client daemonset: %v", err)
	}
	clientDaemonSet.Status.NumberReady = 2
	clientDaemonSet.Status.NumberAvailable = 2
	clientDaemonSet.Status.NumberUnavailable = 0
	if err := client.Status().Update(context.TODO(), clientDaemonSet); err != nil {
		t.Fatalf("failed to update client daemonset status: %v", err)
	}

	ready, err = engine.CheckAndUpdateRuntimeStatus(value)
	if err != nil {
		t.Fatalf("CheckAndUpdateRuntimeStatus() unexpected error = %v", err)
	}
	if !ready {
		t.Fatalf("expected runtime to become ready once all client replicas are ready")
	}

	if err := client.Get(context.TODO(), types.NamespacedName{Name: runtimeName, Namespace: namespace}, updatedRuntime); err != nil {
		t.Fatalf("failed to get runtime after full-ready status update: %v", err)
	}
	if updatedRuntime.Status.Client.Phase != datav1alpha1.RuntimePhaseReady {
		t.Fatalf("expected client phase %q, got %q", datav1alpha1.RuntimePhaseReady, updatedRuntime.Status.Client.Phase)
	}
	if updatedRuntime.Status.SetupDuration == "" {
		t.Fatalf("expected setup duration to be recorded once all client replicas are ready")
	}
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
