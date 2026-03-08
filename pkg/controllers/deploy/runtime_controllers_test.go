/*
Copyright 2022 The Fluid Author.

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

package deploy

import (
	"context"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func restorePrecheckState(t *testing.T) {
	t.Helper()

	precheckFuncsMu.Lock()
	originalResolver := resolveDefaultPrecheckFuncs
	originalPrecheckFuncs := clonePrecheckFuncs(precheckFuncs)
	precheckFuncsMu.Unlock()

	t.Cleanup(func() {
		precheckFuncsMu.Lock()
		defer precheckFuncsMu.Unlock()

		resolveDefaultPrecheckFuncs = originalResolver
		precheckFuncs = clonePrecheckFuncs(originalPrecheckFuncs)
	})
}

func TestGetPrecheckFuncsResolvesDefaultsLazily(t *testing.T) {
	restorePrecheckState(t)

	calls := 0
	check := func(client.Client, types.NamespacedName) (bool, error) {
		return false, nil
	}

	precheckFuncsMu.Lock()
	precheckFuncs = nil
	resolveDefaultPrecheckFuncs = func() map[string]CheckFunc {
		calls++
		return map[string]CheckFunc{"lazy-controller": check}
	}
	precheckFuncsMu.Unlock()

	got := getPrecheckFuncs()
	if calls != 1 {
		t.Fatalf("expected lazy resolver to be called once, got %d", calls)
	}
	if got["lazy-controller"] == nil {
		t.Fatalf("expected lazy resolver result to be returned")
	}

	delete(got, "lazy-controller")
	if getPrecheckFuncs()["lazy-controller"] == nil {
		t.Fatalf("expected getPrecheckFuncs to return a cloned map")
	}
}

func TestScaleoutRuntimeControllerOnDemandUsesInjectedPrecheckFuncsWithoutResolvingDefaults(t *testing.T) {
	restorePrecheckState(t)
	t.Setenv(common.MyPodNamespace, common.NamespaceFluidSystem)

	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-controller",
			Namespace: common.NamespaceFluidSystem,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To[int32](0),
		},
	})

	resolverCalls := 0
	precheckCalls := 0
	precheckFuncsMu.Lock()
	resolveDefaultPrecheckFuncs = func() map[string]CheckFunc {
		resolverCalls++
		return map[string]CheckFunc{}
	}
	precheckFuncsMu.Unlock()
	setPrecheckFunc(map[string]CheckFunc{
		"custom-controller": func(client.Client, types.NamespacedName) (bool, error) {
			precheckCalls++
			return true, nil
		},
	})

	controllerName, scaleout, err := ScaleoutRuntimeControllerOnDemand(fakeClient, types.NamespacedName{
		Namespace: corev1.NamespaceDefault,
		Name:      "dataset",
	}, fake.NullLogger())
	if err != nil {
		t.Fatalf("ScaleoutRuntimeControllerOnDemand() error = %v", err)
	}
	if controllerName != "custom-controller" {
		t.Fatalf("ScaleoutRuntimeControllerOnDemand() controller = %q, want %q", controllerName, "custom-controller")
	}
	if !scaleout {
		t.Fatalf("ScaleoutRuntimeControllerOnDemand() scaleout = false, want true")
	}
	if resolverCalls != 0 {
		t.Fatalf("expected injected prechecks to bypass lazy resolver, got %d resolver calls", resolverCalls)
	}
	if precheckCalls != 1 {
		t.Fatalf("expected injected precheck to run once, got %d", precheckCalls)
	}

	deploy := &appsv1.Deployment{}
	err = fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: common.NamespaceFluidSystem,
		Name:      "custom-controller",
	}, deploy)
	if err != nil {
		t.Fatalf("failed to fetch deployment: %v", err)
	}
	if *deploy.Spec.Replicas != 1 {
		t.Fatalf("expected deployment replicas to scale to 1, got %d", *deploy.Spec.Replicas)
	}
}

func Test_scaleoutDeploymentIfNeeded(t *testing.T) {
	type args struct {
		key types.NamespacedName
		log logr.Logger
	}
	tests := []struct {
		name         string
		args         args
		wantScale    bool
		wantErr      bool
		wantGotErr   bool
		wantReplicas int32
	}{
		// TODO: Add test cases.
		{
			name: "notFound",
			args: args{
				key: types.NamespacedName{
					Namespace: "default",
					Name:      "notFoundController",
				},
				log: fake.NullLogger(),
			}, wantErr: true,
			wantScale:  false,
			wantGotErr: true,
		}, {
			name: "scale to 1 without annotations",
			args: args{
				key: types.NamespacedName{
					Namespace: common.NamespaceFluidSystem,
					Name:      "unknown-Controller",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantScale:    true,
			wantGotErr:   false,
			wantReplicas: 1,
		}, {
			name: "scale to 1 annotations 3",
			args: args{
				key: types.NamespacedName{
					Namespace: common.NamespaceFluidSystem,
					Name:      "goosefsruntime-controller",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantScale:    true,
			wantGotErr:   false,
			wantReplicas: 3,
		}, {
			name: "scale to 1 annotations 0",
			args: args{
				key: types.NamespacedName{
					Namespace: common.NamespaceFluidSystem,
					Name:      "juicefsruntime-controller",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantScale:    true,
			wantGotErr:   false,
			wantReplicas: 1,
		},
	}

	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	deployments := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "alluxioruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jindoruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
				Annotations: map[string]string{
					common.RuntimeControllerReplicas: "0",
				},
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "goosefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
				Annotations: map[string]string{
					common.RuntimeControllerReplicas: "3",
				},
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unknown-Controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		},
	}

	for _, deployment := range deployments {
		objs = append(objs, deployment)
	}

	objs = append(objs, &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: corev1.NamespaceDefault,
		},
	}, &datav1alpha1.GooseFSRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "goosefs",
		Namespace: corev1.NamespaceDefault,
	}}, &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "jindo",
		Namespace: corev1.NamespaceDefault,
	}}, &datav1alpha1.JuiceFSRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "juicefs",
		Namespace: corev1.NamespaceDefault,
	}})

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	setPrecheckFunc(map[string]CheckFunc{
		"alluxioruntime-controller": alluxio.Precheck,
		"jindoruntime-controller":   jindofsx.Precheck,
		"juicefsruntime-controller": juicefs.Precheck,
		"goosefsruntime-controller": goosefs.Precheck,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScale, err := scaleoutDeploymentIfNeeded(fakeClient, tt.args.key, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("scaleoutDeploymentIfNeeded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotScale != tt.wantScale {
				t.Errorf("scaleoutDeploymentIfNeeded() = %v, want %v", gotScale, tt.wantScale)
				return
			}

			deploy := &appsv1.Deployment{}
			err = fakeClient.Get(context.TODO(), tt.args.key, deploy)
			if (err != nil) != tt.wantGotErr {
				t.Errorf("getDeployment() error = %v, wantErr %v", err, tt.wantGotErr)
				return
			}

			if err == nil {
				gotReplicas := *deploy.Spec.Replicas
				if gotReplicas != tt.wantReplicas {
					t.Errorf("scaleoutDeploymentIfNeeded() replicas = %v, want %v", gotReplicas, tt.wantReplicas)
				}
			}

		})
	}
}

func TestScaleoutRuntimeContollerOnDemand(t *testing.T) {
	type args struct {
		key types.NamespacedName
		log logr.Logger
	}
	tests := []struct {
		name               string
		args               args
		wantControllerName string
		wantScaleout       bool
		wantErr            bool
		wantReplicas       int32
	}{
		// TODO: Add test cases.
		{
			name: "notFound",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "notFound",
				},
				log: fake.NullLogger(),
			}, wantErr: true,
			wantControllerName: "",
			wantScaleout:       false,
		}, {
			name: "unknown",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "unknown",
				},
				log: fake.NullLogger(),
			},
			wantErr:            true,
			wantControllerName: "",
			wantScaleout:       false,
		}, {
			name: "scale alluxio runtime to 1 without annotations",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "alluxio",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantControllerName: "alluxioruntime-controller",
			wantScaleout:       true,
			wantReplicas:       1,
		}, {
			name: "no need to scale jindo runtime",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "jindo",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantControllerName: "jindoruntime-controller",
			wantScaleout:       false,
			wantReplicas:       1,
		}, {
			name: "scale juice runtime with annotation 0",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "juicefs",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantControllerName: "juicefsruntime-controller",
			wantScaleout:       true,
			wantReplicas:       1,
		}, {
			name: "scale goosef runtime with annotation 0",
			args: args{
				key: types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      "goosefs",
				},
				log: fake.NullLogger(),
			}, wantErr: false,
			wantControllerName: "goosefsruntime-controller",
			wantScaleout:       true,
			wantReplicas:       3,
		},
	}

	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	deployments := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "alluxioruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jindoruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
				Annotations: map[string]string{
					common.RuntimeControllerReplicas: "0",
				},
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "goosefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
				Annotations: map[string]string{
					common.RuntimeControllerReplicas: "3",
				},
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unknown-Controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](0),
			},
		},
	}

	for _, deployment := range deployments {
		objs = append(objs, deployment)
	}

	objs = append(objs, &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: corev1.NamespaceDefault,
		},
	}, &datav1alpha1.GooseFSRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "goosefs",
		Namespace: corev1.NamespaceDefault,
	}}, &datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "jindo",
		Namespace: corev1.NamespaceDefault,
	}}, &datav1alpha1.JuiceFSRuntime{ObjectMeta: metav1.ObjectMeta{
		Name:      "juicefs",
		Namespace: corev1.NamespaceDefault,
	}})

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(common.MyPodNamespace, common.NamespaceFluidSystem)
			gotControllerName, gotScaleout, err := ScaleoutRuntimeControllerOnDemand(fakeClient, tt.args.key, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScaleoutRuntimeControllerOnDemand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotControllerName != tt.wantControllerName {
				t.Errorf("ScaleoutRuntimeControllerOnDemand() gotControllerName = %v, want %v", gotControllerName, tt.wantControllerName)
			}
			if gotScaleout != tt.wantScaleout {
				t.Errorf("ScaleoutRuntimeControllerOnDemand() gotScaleout = %v, want %v", gotScaleout, tt.wantScaleout)
			}

			if tt.wantControllerName != "" {
				deploy := &appsv1.Deployment{}
				err = fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: common.NamespaceFluidSystem,
					Name:      tt.wantControllerName,
				}, deploy)
				if err != nil {
					t.Errorf("getDeployment() error = %v", err)
					return
				}

				if err == nil {
					gotReplicas := *deploy.Spec.Replicas
					if gotReplicas != tt.wantReplicas {
						t.Errorf("scaleoutDeploymentIfNeeded() replicas = %v, want %v", gotReplicas, tt.wantReplicas)
					}
				}
			}
		})
	}
}
