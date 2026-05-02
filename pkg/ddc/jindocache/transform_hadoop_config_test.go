package jindocache

import (
	. "github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var _ = Describe("JindoCacheEngine transformHadoopConfig", func() {
	newEngine := func(objects ...runtime.Object) *JindoCacheEngine {
		return &JindoCacheEngine{Client: NewFakeClientWithScheme(testScheme, objects...)}
	}

	newRuntime := func(configMapName string) *datav1alpha1.JindoRuntime {
		return &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "fluid"},
			Spec: datav1alpha1.JindoRuntimeSpec{
				HadoopConfig: configMapName,
			},
		}
	}

	It("should skip lookup when no hadoop config is configured", func() {
		engine := newEngine()
		runtime := newRuntime("")
		value := &Jindo{}

		err := engine.transformHadoopConfig(runtime, value)

		Expect(err).NotTo(HaveOccurred())
		Expect(value.HadoopConfig).To(Equal(HadoopConfig{}))
	})

	It("should report a missing configmap", func() {
		engine := newEngine()
		runtime := newRuntime("missing-hadoop-config")

		err := engine.transformHadoopConfig(runtime, &Jindo{})

		Expect(err).To(MatchError("specified hadoopConfig \"missing-hadoop-config\" is not found"))
	})

	It("should reject configmaps that do not contain core-site.xml or hdfs-site.xml", func() {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "invalid-hadoop-config", Namespace: "fluid"},
			Data: map[string]string{
				"unrelated.xml": "<configuration />",
			},
		}
		engine := newEngine(configMap)
		runtime := newRuntime(configMap.Name)

		err := engine.transformHadoopConfig(runtime, &Jindo{})

		Expect(err).To(MatchError("neither \"hdfs-site.xml\" nor \"core-site.xml\" is found in the specified configMap \"invalid-hadoop-config\" "))
	})

	It("should include both supported hadoop config files when present", func() {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "valid-hadoop-config", Namespace: "fluid"},
			Data: map[string]string{
				HADOOP_CONF_HDFS_SITE_FILENAME: "<configuration />",
				HADOOP_CONF_CORE_SITE_FILENAME: "<configuration />",
			},
		}
		engine := newEngine(configMap)
		runtime := newRuntime(configMap.Name)
		value := &Jindo{}

		err := engine.transformHadoopConfig(runtime, value)

		Expect(err).NotTo(HaveOccurred())
		Expect(value.HadoopConfig.ConfigMap).To(Equal(configMap.Name))
		Expect(value.HadoopConfig.IncludeHdfsSite).To(BeTrue())
		Expect(value.HadoopConfig.IncludeCoreSite).To(BeTrue())
	})
})
