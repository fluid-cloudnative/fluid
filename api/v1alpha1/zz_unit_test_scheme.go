// zz_unit_test_scheme.go with a "zz" prefix to ensure its init function is called after all the other init function in the pacakge
package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	UnitTestScheme *runtime.Scheme
)

func init() {
	UnitTestScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(UnitTestScheme)
	_ = AddToScheme(UnitTestScheme)
	_ = appsv1.AddToScheme(UnitTestScheme)
}
