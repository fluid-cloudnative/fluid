package componenthelper

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

const (
	StatefulSetWorkloadType  string = "apps/v1/StatefulSet"
	DaemonsetSetWorkloadType string = "apps/v1/DaemonSet"
)

type ComponentHelper interface {
	Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error
	CheckComponentExist(ctx context.Context, component *common.CacheRuntimeComponentValue) (bool, error)
	ConstructComponentStatus(ctx context.Context, component *common.CacheRuntimeComponentValue) (datav1alpha1.RuntimeComponentStatus, error)
	GetComponentTopologyInfo(ctx context.Context, component *common.CacheRuntimeComponentValue) (common.TopologyConfig, error)
	CleanupOrphanedComponentResources(ctx context.Context, component *common.CacheRuntimeComponentValue) error
}

func NewComponentHelper(workloadType metav1.TypeMeta, scheme *runtime.Scheme, client client.Client) ComponentHelper {
	if workloadType.APIVersion == "apps/v1" {
		if workloadType.Kind == "StatefulSet" {
			return NewStatefulSetReconciler(scheme, client)
		} else if workloadType.Kind == "DaemonSet" {
			return NewDaemonsetReconciler(scheme, client)
		}
	}

	return NewStatefulSetReconciler(scheme, client)
}
