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

package alluxio

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

type erroringGetClient struct {
	client.Client
	err error
}

func (c erroringGetClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return c.err
}

var _ = Describe("Alluxio engine and transform file scope", Label("pkg.ddc.alluxio.engine_transform_file_scope_test.go"), func() {
	Describe("Build", func() {
		It("returns an error when runtime is missing", func() {
			engine, err := Build("test-id", cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{Name: "demo", Namespace: "fluid"},
				Log:            fake.NullLogger(),
				RuntimeType:    "alluxio",
			})

			Expect(err).To(MatchError("engine demo is failed to parse"))
			Expect(engine).To(BeNil())
		})

		It("returns an error when runtime has the wrong type", func() {
			engine, err := Build("test-id", cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{Name: "demo", Namespace: "fluid"},
				Log:            fake.NullLogger(),
				RuntimeType:    "alluxio",
				Runtime:        &datav1alpha1.ThinRuntime{},
			})

			Expect(err).To(MatchError("engine demo is failed to parse"))
			Expect(engine).To(BeNil())
		})
	})

	Describe("Precheck", func() {
		It("reports whether the alluxio runtime exists", func() {
			key := types.NamespacedName{Name: "demo", Namespace: "fluid"}
			runtime := &datav1alpha1.AlluxioRuntime{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}}
			client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, runtime)

			found, err := Precheck(client, key)

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
		})

		It("returns false without error when the runtime is absent", func() {
			found, err := Precheck(fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme), types.NamespacedName{Name: "missing", Namespace: "fluid"})

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("returns the client error for non-not-found failures", func() {
			failingClient := erroringGetClient{
				Client: fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme),
				err:    apierrors.NewForbidden(schema.GroupResource{Group: datav1alpha1.GroupVersion.Group, Resource: "alluxioruntimes"}, "demo", nil),
			}

			found, err := Precheck(failingClient, types.NamespacedName{Name: "demo", Namespace: "fluid"})

			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
			Expect(found).To(BeFalse())
		})
	})

	Describe("transformPlacementMode", func() {
		It("defaults to exclusive placement when the dataset placement mode is empty", func() {
			value := &Alluxio{}

			(&AlluxioEngine{}).transformPlacementMode(&datav1alpha1.Dataset{}, value)

			Expect(value.PlacementMode).To(Equal(string(datav1alpha1.ExclusiveMode)))
		})
	})

	Describe("transform", func() {
		It("builds an alluxio value for a dataset-backed runtime", func() {
			namespacedName := types.NamespacedName{Name: "demo", Namespace: "fluid"}
			dataset, runtime := mockFluidObjectsForTests(namespacedName)
			dataset.Spec.PlacementMode = ""
			runtime.Spec.AlluxioVersion.ImageTag = "2.8.0"
			runtime.Spec.Master.NetworkMode = datav1alpha1.ContainerNetworkMode
			runtime.Spec.Worker.NetworkMode = datav1alpha1.ContainerNetworkMode
			runtime.Spec.APIGateway.Enabled = true
			runtime.Spec.HadoopConfig = "demo-hadoop-conf"
			runtime.Spec.PodMetadata = datav1alpha1.PodMetadata{
				Labels:      map[string]string{"common-label": "common-value"},
				Annotations: map[string]string{"common-annotation": "common-value"},
			}
			dataset.Spec.Mounts = []datav1alpha1.Mount{{
				MountPoint: "https://downloads.example.com/demo",
				Path:       "/",
				Name:       "http",
			}}

			runtimeInfo, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, "alluxio", base.WithTieredStore(runtime.Spec.TieredStore))
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetOwnerDatasetUID(dataset.UID)
			runtimeInfo.SetupWithDataset(dataset)

			engine := mockAlluxioEngineForTests(dataset, runtime)
			engine.runtimeInfo = runtimeInfo
			engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme,
				dataset,
				runtime,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: runtime.Spec.HadoopConfig, Namespace: runtime.Namespace},
					Data: map[string]string{
						hadoopConfCoreSiteFilename: "core-site",
						hadoopConfHdfsSiteFilename: "hdfs-site",
					},
				},
			)

			value, err := engine.transform(runtime)

			Expect(err).NotTo(HaveOccurred())
			Expect(value).NotTo(BeNil())
			Expect(value.FullnameOverride).To(Equal(namespacedName.Name))
			Expect(value.OwnerDatasetId).To(Equal(utils.GetDatasetId(namespacedName.Namespace, namespacedName.Name, string(dataset.UID))))
			Expect(value.Properties["alluxio.master.mount.table.root.ufs"]).To(Equal("https://downloads.example.com/demo"))
			Expect(value.Properties["alluxio.user.file.replication.max"]).To(Equal("1"))
			Expect(value.Properties["alluxio.proxy.web.port"]).To(Equal("39999"))
			Expect(value.Properties["alluxio.underfs.hdfs.configuration"]).To(ContainSubstring(hadoopConfMountPath + "/" + hadoopConfCoreSiteFilename))
			Expect(value.Properties["alluxio.underfs.hdfs.configuration"]).To(ContainSubstring(hadoopConfMountPath + "/" + hadoopConfHdfsSiteFilename))
			Expect(value.Properties["alluxio.user.block.size.bytes.default"]).To(Equal("256MB"))
			Expect(value.Monitoring).To(Equal(alluxioRuntimeMetricsLabel))
			Expect(value.Master.Labels).To(HaveKeyWithValue("common-label", "common-value"))
			Expect(value.Worker.Annotations).To(HaveKeyWithValue("common-annotation", "common-value"))
			Expect(value.APIGateway.Enabled).To(BeTrue())
			Expect(value.Fuse.Args).To(HaveLen(2))
			Expect(value.Fuse.Args[1]).To(ContainSubstring("max_readahead=0"))
			Expect(value.Fuse.Args[1]).To(ContainSubstring("allow_other"))
			Expect(value.Fuse.NodeSelector).To(HaveKeyWithValue(utils.GetFuseLabelName(runtime.Namespace, runtime.Name, string(dataset.UID)), "true"))
			Expect(value.TieredStore.Levels).To(HaveLen(1))
			Expect(value.TieredStore.Levels[0].Path).To(Equal("/dev/shm/fluid/demo"))
			Expect(value.PlacementMode).To(Equal(string(datav1alpha1.ExclusiveMode)))
		})
	})
})
