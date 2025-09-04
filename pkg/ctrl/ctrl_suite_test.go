package ctrl

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCtrl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ctrl Suite")
}

func mockRuntimeStatefulset(name, namespace string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec:   appsv1.StatefulSetSpec{},
		Status: appsv1.StatefulSetStatus{},
	}
}

func mockRuntimeDaemonset(name, namespace string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec:   appsv1.DaemonSetSpec{},
		Status: appsv1.DaemonSetStatus{},
	}
}
