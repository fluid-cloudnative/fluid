/*
Copyright 2022 The Fluid Authors.

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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestAlluxioEngine_SyncRuntime(t *testing.T) {
	t.Setenv(common.EnvImagePullSecretsKey, "")

	const (
		name      = "test-runtime"
		namespace = "default"
	)
	commonSecret := []corev1.LocalObjectReference{{Name: "common-secret"}}
	masterSecret := []corev1.LocalObjectReference{{Name: "master-secret"}}
	fuseSecret := []corev1.LocalObjectReference{{Name: "fuse-secret"}}
	oldSecret := []corev1.LocalObjectReference{{Name: "old-secret"}}

	alluxioRuntime := &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			ImagePullSecrets: commonSecret,
			Master: datav1alpha1.AlluxioCompTemplateSpec{
				ImagePullSecrets: masterSecret,
			},
			Fuse: datav1alpha1.AlluxioFuseSpec{
				ImagePullSecrets: fuseSecret,
			},
		},
	}
	master := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-master", Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				ImagePullSecrets: oldSecret,
				Containers:       []corev1.Container{{Name: "master"}},
			}},
		},
	}
	worker := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-worker", Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				ImagePullSecrets: oldSecret,
				Containers:       []corev1.Container{{Name: "worker"}},
			}},
		},
	}
	fuse := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-fuse", Namespace: namespace},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				ImagePullSecrets: oldSecret,
				Containers:       []corev1.Container{{Name: "fuse"}},
			}},
		},
	}

	client := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime, master, worker, fuse)
	engine := &AlluxioEngine{
		name:      name,
		namespace: namespace,
		Client:    client,
		Log:       fake.NullLogger(),
	}
	ctx := runtime.ReconcileRequestContext{}

	changed, err := engine.SyncRuntime(ctx)
	if err != nil {
		t.Fatalf("SyncRuntime() error = %v", err)
	}
	if !changed {
		t.Fatal("SyncRuntime() changed = false, want true")
	}

	checkSecrets := func(wantMaster, wantWorker, wantFuse []corev1.LocalObjectReference) {
		t.Helper()
		gotMaster := &appsv1.StatefulSet{}
		gotWorker := &appsv1.StatefulSet{}
		gotFuse := &appsv1.DaemonSet{}
		if err := client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: master.Name}, gotMaster); err != nil {
			t.Fatalf("failed to get master StatefulSet: %v", err)
		}
		if err := client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: worker.Name}, gotWorker); err != nil {
			t.Fatalf("failed to get worker StatefulSet: %v", err)
		}
		if err := client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: fuse.Name}, gotFuse); err != nil {
			t.Fatalf("failed to get fuse DaemonSet: %v", err)
		}

		if !reflect.DeepEqual(gotMaster.Spec.Template.Spec.ImagePullSecrets, wantMaster) {
			t.Errorf("master imagePullSecrets = %v, want %v", gotMaster.Spec.Template.Spec.ImagePullSecrets, wantMaster)
		}
		if !reflect.DeepEqual(gotWorker.Spec.Template.Spec.ImagePullSecrets, wantWorker) {
			t.Errorf("worker imagePullSecrets = %v, want %v", gotWorker.Spec.Template.Spec.ImagePullSecrets, wantWorker)
		}
		if !reflect.DeepEqual(gotFuse.Spec.Template.Spec.ImagePullSecrets, wantFuse) {
			t.Errorf("fuse imagePullSecrets = %v, want %v", gotFuse.Spec.Template.Spec.ImagePullSecrets, wantFuse)
		}
		if len(gotMaster.Spec.Template.Spec.Containers) != 1 || gotMaster.Spec.Template.Spec.Containers[0].Name != "master" {
			t.Errorf("master containers were changed: %v", gotMaster.Spec.Template.Spec.Containers)
		}
		if len(gotWorker.Spec.Template.Spec.Containers) != 1 || gotWorker.Spec.Template.Spec.Containers[0].Name != "worker" {
			t.Errorf("worker containers were changed: %v", gotWorker.Spec.Template.Spec.Containers)
		}
		if len(gotFuse.Spec.Template.Spec.Containers) != 1 || gotFuse.Spec.Template.Spec.Containers[0].Name != "fuse" {
			t.Errorf("fuse containers were changed: %v", gotFuse.Spec.Template.Spec.Containers)
		}
	}
	checkSecrets(masterSecret, commonSecret, fuseSecret)

	changed, err = engine.SyncRuntime(ctx)
	if err != nil {
		t.Fatalf("SyncRuntime() on unchanged runtime error = %v", err)
	}
	if changed {
		t.Fatal("SyncRuntime() on unchanged runtime changed = true, want false")
	}

	rotatedCommonSecret := []corev1.LocalObjectReference{{Name: "rotated-common-secret"}}
	alluxioRuntime.Spec.ImagePullSecrets = rotatedCommonSecret
	alluxioRuntime.Spec.Master.ImagePullSecrets = nil
	alluxioRuntime.Spec.Worker.ImagePullSecrets = nil
	alluxioRuntime.Spec.Fuse.ImagePullSecrets = nil
	if err := client.Update(context.Background(), alluxioRuntime); err != nil {
		t.Fatalf("failed to rotate common runtime imagePullSecrets: %v", err)
	}
	changed, err = engine.SyncRuntime(ctx)
	if err != nil {
		t.Fatalf("SyncRuntime() while rotating common secrets error = %v", err)
	}
	if !changed {
		t.Fatal("SyncRuntime() while rotating common secrets changed = false, want true")
	}
	checkSecrets(rotatedCommonSecret, rotatedCommonSecret, rotatedCommonSecret)

	alluxioRuntime.Spec.ImagePullSecrets = nil
	if err := client.Update(context.Background(), alluxioRuntime); err != nil {
		t.Fatalf("failed to remove runtime imagePullSecrets: %v", err)
	}
	changed, err = engine.SyncRuntime(ctx)
	if err != nil {
		t.Fatalf("SyncRuntime() while removing secrets error = %v", err)
	}
	if !changed {
		t.Fatal("SyncRuntime() while removing secrets changed = false, want true")
	}
	checkSecrets(nil, nil, nil)
}
