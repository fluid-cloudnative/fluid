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

package thin

import (
	"context"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ThinEngine genFuseMountOptions", func() {
	newEngine := func(objs ...k8sruntime.Object) *ThinEngine {
		return &ThinEngine{
			name:      "dataset",
			namespace: "fluid",
			Client:    fake.NewFakeClientWithScheme(testScheme, objs...),
			Log:       fake.NullLogger(),
		}
	}

	newSecret := func() *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "creds",
				Namespace: "fluid",
			},
			Data: map[string][]byte{
				"shared-key": []byte("shared-secret"),
				"mount-key":  []byte("mount-secret"),
			},
		}
	}

	It("merges mount options with shared and mount encrypt options", func() {
		engine := newEngine(newSecret())
		mount := datav1alpha1.Mount{
			Name:    "demo-mount",
			Options: map[string]string{"fs.mount": "mount-option"},
			EncryptOptions: []datav1alpha1.EncryptOption{{
				Name: "fs.mount.secret",
				ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
					Name: "creds",
					Key:  "mount-key",
				}},
			}},
		}

		options, err := engine.genFuseMountOptions(
			mount,
			map[string]string{"fs.shared": "shared-option"},
			[]datav1alpha1.EncryptOption{{
				Name: "fs.shared.secret",
				ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
					Name: "creds",
					Key:  "shared-key",
				}},
			}},
			true,
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(map[string]string{
			"fs.shared":        "shared-option",
			"fs.mount":         "mount-option",
			"fs.shared.secret": "shared-secret",
			"fs.mount.secret":  "mount-secret",
		}))
	})

	It("skips encrypt option extraction when disabled", func() {
		engine := newEngine(newSecret())
		mount := datav1alpha1.Mount{
			Name:           "demo-mount",
			Options:        map[string]string{"fs.mount": "mount-option"},
			EncryptOptions: []datav1alpha1.EncryptOption{{Name: "fs.mount.secret"}},
		}

		options, err := engine.genFuseMountOptions(
			mount,
			map[string]string{"fs.shared": "shared-option"},
			[]datav1alpha1.EncryptOption{{Name: "fs.shared.secret"}},
			false,
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(map[string]string{
			"fs.shared": "shared-option",
			"fs.mount":  "mount-option",
		}))
	})

	It("returns an error when an encrypt option duplicates an existing option", func() {
		engine := newEngine(newSecret())
		mount := datav1alpha1.Mount{
			Name:    "demo-mount",
			Options: map[string]string{"fs.mount.secret": "already-set"},
			EncryptOptions: []datav1alpha1.EncryptOption{{
				Name: "fs.mount.secret",
				ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
					Name: "creds",
					Key:  "mount-key",
				}},
			}},
		}

		_, err := engine.genFuseMountOptions(mount, nil, nil, true)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("set more than one times"))
	})
})

var _ = Describe("ThinEngine updateFusePod", func() {
	newEngine := func(objs ...k8sruntime.Object) *ThinEngine {
		return &ThinEngine{
			name:      "dataset",
			namespace: "fluid",
			Client:    fake.NewFakeClientWithScheme(testScheme, objs...),
			Log:       fake.NullLogger(),
		}
	}

	newFuseDaemonSet := func() *appsv1.DaemonSet {
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "fuse"}},
			},
		}
	}

	newFusePod := func(name string, annotations map[string]string, ready bool) *corev1.Pod {
		conditionStatus := corev1.ConditionFalse
		if ready {
			conditionStatus = corev1.ConditionTrue
		}

		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   "fluid",
				Labels:      map[string]string{"app": "fuse"},
				Annotations: annotations,
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodReady,
					Status: conditionStatus,
				}},
			},
		}
	}

	annotationKey := common.LabelAnnotationFusePrefix + "update-fuse-config"

	It("adds and increments the fuse config annotation on running pods", func() {
		engine := newEngine(
			newFuseDaemonSet(),
			newFusePod("fuse-0", map[string]string{"existing": "value"}, true),
			newFusePod("fuse-1", map[string]string{annotationKey: "3"}, true),
			newFusePod("fuse-2", map[string]string{annotationKey: "9"}, false),
		)

		Expect(engine.updateFusePod()).To(Succeed())

		updatedReadyPod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "fuse-0", Namespace: "fluid"}, updatedReadyPod)).To(Succeed())
		Expect(updatedReadyPod.Annotations).To(HaveKeyWithValue(annotationKey, "1"))

		updatedIncrementedPod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "fuse-1", Namespace: "fluid"}, updatedIncrementedPod)).To(Succeed())
		Expect(updatedIncrementedPod.Annotations).To(HaveKeyWithValue(annotationKey, "4"))

		notReadyPod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "fuse-2", Namespace: "fluid"}, notReadyPod)).To(Succeed())
		Expect(notReadyPod.Annotations).To(HaveKeyWithValue(annotationKey, "9"))
	})

	It("returns nil when no running fuse pod is found", func() {
		engine := newEngine(newFuseDaemonSet())

		Expect(engine.updateFusePod()).To(Succeed())
	})

	It("returns an error when an existing annotation is not numeric", func() {
		engine := newEngine(
			newFuseDaemonSet(),
			newFusePod("fuse-0", map[string]string{annotationKey: "invalid"}, true),
		)

		err := engine.updateFusePod()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid syntax"))

		pod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "fuse-0", Namespace: "fluid"}, pod)).To(Succeed())
		Expect(pod.Annotations).To(HaveKeyWithValue(annotationKey, "invalid"))
	})
})

var _ = Describe("ThinEngine ShouldUpdateUFS", func() {
	newEngine := func(runtime *datav1alpha1.ThinRuntime, objs ...k8sruntime.Object) *ThinEngine {
		allObjects := append([]k8sruntime.Object{runtime}, objs...)
		return &ThinEngine{
			name:      runtime.Name,
			namespace: runtime.Namespace,
			runtime:   runtime,
			Client:    fake.NewFakeClientWithScheme(testScheme, allObjects...),
			Log:       fake.NullLogger(),
		}
	}

	newRuntime := func(name string) *datav1alpha1.ThinRuntime {
		return &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fluid",
			},
		}
	}

	newDataset := func(name string, mounts ...datav1alpha1.Mount) *datav1alpha1.Dataset {
		return &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{Mounts: mounts},
		}
	}

	newFuseConfigMap := func(name, config string) *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name + "-fuse-conf",
				Namespace: "fluid",
			},
			Data: map[string]string{"config.json": config},
		}
	}

	newFuseDaemonSet := func(name string) *appsv1.DaemonSet {
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name + "-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name + "-fuse"}},
			},
		}
	}

	newFusePod := func(runtimeName, podName string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        podName,
				Namespace:   "fluid",
				Labels:      map[string]string{"app": runtimeName + "-fuse"},
				Annotations: map[string]string{"existing": "value"},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				}},
			},
		}
	}

	BeforeEach(func() {
		Expect(os.Setenv(EnvFuseConfigStorage, "configmap")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Unsetenv(EnvFuseConfigStorage)).To(Succeed())
	})

	It("updates the fuse configmap and nudges fuse pods when dataset mounts change", func() {
		runtime := newRuntime("sync")
		dataset := newDataset(
			"sync",
			datav1alpha1.Mount{Name: "zookeeper", MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/stable/"},
			datav1alpha1.Mount{Name: "hbase", MountPoint: "https://mirrors.bit.edu.cn/apache/hbase/stable/"},
		)
		configMap := newFuseConfigMap("sync", "{\"mounts\":[{\"mountPoint\":\"https://mirrors.bit.edu.cn/apache/zookeeper/stable/\",\"name\":\"zookeeper\"}],\"targetPath\":\"/thin/fluid/sync/thin-fuse\",\"accessModes\":[\"ReadOnlyMany\"]}")
		engine := newEngine(runtime, dataset, configMap, newFuseDaemonSet("sync"), newFusePod("sync", "sync-fuse-0"))

		Expect(engine.ShouldUpdateUFS()).To(BeNil())

		updatedConfigMap := &corev1.ConfigMap{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "sync-fuse-conf", Namespace: "fluid"}, updatedConfigMap)).To(Succeed())
		Expect(updatedConfigMap.Data["config.json"]).To(ContainSubstring("apache/hbase/stable"))

		updatedPod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "sync-fuse-0", Namespace: "fluid"}, updatedPod)).To(Succeed())
		Expect(updatedPod.Annotations).To(HaveKeyWithValue(common.LabelAnnotationFusePrefix+"update-fuse-config", "1"))
	})

	It("leaves fuse pods untouched when the generated config is unchanged", func() {
		runtime := newRuntime("steady")
		dataset := newDataset(
			"steady",
			datav1alpha1.Mount{Name: "zookeeper", MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/stable/"},
		)
		config := "{\"mounts\":[{\"mountPoint\":\"https://mirrors.bit.edu.cn/apache/zookeeper/stable/\",\"name\":\"zookeeper\"}],\"targetPath\":\"/thin/fluid/steady/thin-fuse\",\"accessModes\":[\"ReadOnlyMany\"]}"
		engine := newEngine(runtime, dataset, newFuseConfigMap("steady", config), newFuseDaemonSet("steady"), newFusePod("steady", "steady-fuse-0"))

		Expect(engine.ShouldUpdateUFS()).To(BeNil())

		pod := &corev1.Pod{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "steady-fuse-0", Namespace: "fluid"}, pod)).To(Succeed())
		Expect(pod.Annotations).NotTo(HaveKey(common.LabelAnnotationFusePrefix + "update-fuse-config"))
	})

	It("returns nil when the dataset is not found", func() {
		runtime := newRuntime("missing")
		engine := newEngine(runtime)

		Expect(engine.ShouldUpdateUFS()).To(BeNil())
	})
})
