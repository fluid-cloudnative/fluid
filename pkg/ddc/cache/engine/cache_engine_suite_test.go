package engine

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CacheEngineTestScheme is the scheme used for cache engine tests
var CacheEngineTestScheme *runtime.Scheme

func init() {
	CacheEngineTestScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(CacheEngineTestScheme)
	_ = appsv1.AddToScheme(CacheEngineTestScheme)
	_ = datav1alpha1.AddToScheme(CacheEngineTestScheme)
	_ = workloadv1alpha1.AddToScheme(CacheEngineTestScheme)
}

func TestCacheEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Engine Suite", Label("cache_engine"))
}
