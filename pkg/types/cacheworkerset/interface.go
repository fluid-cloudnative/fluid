package cacheworkerset

import (
	"context"
	"fmt"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	openkruise "github.com/openkruise/kruise/apis/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkerType defines the type of worker resource
type WorkerType string
type StatefulSetConditionType string

var NeedScaleInAnnoKey string = "fluid.io/need-scale-in"

const (
	WorkerTypeAnnoKey       WorkerType = "fluid.io/workerset-type"
	SpecifiedDeleteKey      string     = "apps.kruise.io/specified-delete"
	StatefulSetType         WorkerType = "statefulset"
	AdvancedStatefulSetType WorkerType = "advanced_statefulset"
	DaemonSetType           WorkerType = "daemonset"
)

type CacheWorkerSetInterface interface {
	ToStatefulSet() *appsv1.StatefulSet
	ToDaemonSet() *appsv1.DaemonSet
	ToAdvancedStatefulSet() *openkruise.StatefulSet
}

// CacheWorkerManagerClass defines the manager class
type CacheWorkerSet struct {
	client     client.Client
	WorkerType WorkerType
	Sts        *appsv1.StatefulSet
	Ds         *appsv1.DaemonSet
	Asts       *openkruise.StatefulSet
}

// GetNamespace 根据 WorkerType 获取命名空间
func (c *CacheWorkerSet) GetNamespace() string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Namespace
	case DaemonSetType:
		return c.Ds.Namespace
	case AdvancedStatefulSetType:
		return c.Asts.Namespace
	default:
		return ""
	}
}

// SetNamespace 设置命名空间
func (c *CacheWorkerSet) SetNamespace(namespace string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Namespace = namespace
	case DaemonSetType:
		c.Ds.Namespace = namespace
	case AdvancedStatefulSetType:
		c.Asts.Namespace = namespace
	}
}

// GetName 获取名称
func (c *CacheWorkerSet) GetName() string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Name
	case DaemonSetType:
		return c.Ds.Name
	case AdvancedStatefulSetType:
		return c.Asts.Name
	default:
		return ""
	}
}

// SetName 设置名称
func (c *CacheWorkerSet) SetName(name string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Name = name
	case DaemonSetType:
		c.Ds.Name = name
	case AdvancedStatefulSetType:
		c.Asts.Name = name
	}
}

func (c *CacheWorkerSet) GetGenerateName() string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.GenerateName
	case DaemonSetType:
		return c.Ds.GenerateName
	case AdvancedStatefulSetType:
		return c.Asts.GenerateName
	default:
		return ""
	}
}

func (c *CacheWorkerSet) SetGenerateName(name string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.GenerateName = name
	case DaemonSetType:
		c.Ds.GenerateName = name
	case AdvancedStatefulSetType:
		c.Asts.GenerateName = name
	}
}

// GetUID 获取 UID
func (c *CacheWorkerSet) GetUID() types.UID {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.UID
	case DaemonSetType:
		return c.Ds.UID
	case AdvancedStatefulSetType:
		return c.Asts.UID
	default:
		return ""
	}
}

// SetUID 设置 UID
func (c *CacheWorkerSet) SetUID(uid types.UID) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.UID = uid
	case DaemonSetType:
		c.Ds.UID = uid
	case AdvancedStatefulSetType:
		c.Asts.UID = uid
	}
}

func (c *CacheWorkerSet) GetResourceVersion() string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.ResourceVersion
	case DaemonSetType:
		return c.Ds.ResourceVersion
	case AdvancedStatefulSetType:
		return c.Asts.ResourceVersion
	default:
		return ""
	}
}

func (c *CacheWorkerSet) SetResourceVersion(version string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.ResourceVersion = version
	case DaemonSetType:
		c.Ds.ResourceVersion = version
	case AdvancedStatefulSetType:
		c.Asts.ResourceVersion = version
	}
}

func (c *CacheWorkerSet) GetGeneration() int64 {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Generation
	case DaemonSetType:
		return c.Ds.Generation
	case AdvancedStatefulSetType:
		return c.Asts.Generation
	default:
		return 0
	}
}

func (c *CacheWorkerSet) SetGeneration(generation int64) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Generation = generation
	case DaemonSetType:
		c.Ds.Generation = generation
	case AdvancedStatefulSetType:
		c.Asts.Generation = generation
	}
}

func (c *CacheWorkerSet) GetSelfLink() string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.SelfLink
	case DaemonSetType:
		return c.Ds.SelfLink
	case AdvancedStatefulSetType:
		return c.Asts.SelfLink
	default:
		return ""
	}
}

func (c *CacheWorkerSet) SetSelfLink(selfLink string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.SelfLink = selfLink
	case DaemonSetType:
		c.Ds.SelfLink = selfLink
	case AdvancedStatefulSetType:
		c.Asts.SelfLink = selfLink
	}
}

func (c *CacheWorkerSet) GetCreationTimestamp() metav1.Time {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.CreationTimestamp
	case DaemonSetType:
		return c.Ds.CreationTimestamp
	case AdvancedStatefulSetType:
		return c.Asts.CreationTimestamp
	default:
		return metav1.Time{}
	}
}

func (c *CacheWorkerSet) SetCreationTimestamp(timestamp metav1.Time) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.CreationTimestamp = timestamp
	case DaemonSetType:
		c.Ds.CreationTimestamp = timestamp
	case AdvancedStatefulSetType:
		c.Asts.CreationTimestamp = timestamp
	}
}

func (c *CacheWorkerSet) GetDeletionTimestamp() *metav1.Time {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.DeletionTimestamp
	case DaemonSetType:
		return c.Ds.DeletionTimestamp
	case AdvancedStatefulSetType:
		return c.Asts.DeletionTimestamp
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetDeletionTimestamp(timestamp *metav1.Time) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.DeletionTimestamp = timestamp
	case DaemonSetType:
		c.Ds.DeletionTimestamp = timestamp
	case AdvancedStatefulSetType:
		c.Asts.DeletionTimestamp = timestamp
	}
}

func (c *CacheWorkerSet) GetDeletionGracePeriodSeconds() *int64 {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.DeletionGracePeriodSeconds
	case DaemonSetType:
		return c.Ds.DeletionGracePeriodSeconds
	case AdvancedStatefulSetType:
		return c.Asts.DeletionGracePeriodSeconds
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetDeletionGracePeriodSeconds(i *int64) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.DeletionGracePeriodSeconds = i
	case DaemonSetType:
		c.Ds.DeletionGracePeriodSeconds = i
	case AdvancedStatefulSetType:
		c.Asts.DeletionGracePeriodSeconds = i
	}
}

func (c *CacheWorkerSet) GetLabels() map[string]string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Labels
	case DaemonSetType:
		return c.Ds.Labels
	case AdvancedStatefulSetType:
		return c.Asts.Labels
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetLabels(labels map[string]string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Labels = labels
	case DaemonSetType:
		c.Ds.Labels = labels
	case AdvancedStatefulSetType:
		c.Asts.Labels = labels
	}
}

func (c *CacheWorkerSet) GetAnnotations() map[string]string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Annotations
	case DaemonSetType:
		return c.Ds.Annotations
	case AdvancedStatefulSetType:
		return c.Asts.Annotations
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetAnnotations(annotations map[string]string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Annotations = annotations
	case DaemonSetType:
		c.Ds.Annotations = annotations
	case AdvancedStatefulSetType:
		c.Asts.Annotations = annotations
	}
}

func (c *CacheWorkerSet) GetFinalizers() []string {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.Finalizers
	case DaemonSetType:
		return c.Ds.Finalizers
	case AdvancedStatefulSetType:
		return c.Asts.Finalizers
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetFinalizers(finalizers []string) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.Finalizers = finalizers
	case DaemonSetType:
		c.Ds.Finalizers = finalizers
	case AdvancedStatefulSetType:
		c.Asts.Finalizers = finalizers
	}
}

func (c *CacheWorkerSet) GetOwnerReferences() []metav1.OwnerReference {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.OwnerReferences
	case DaemonSetType:
		return c.Ds.OwnerReferences
	case AdvancedStatefulSetType:
		return c.Asts.OwnerReferences
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetOwnerReferences(references []metav1.OwnerReference) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.OwnerReferences = references
	case DaemonSetType:
		c.Ds.OwnerReferences = references
	case AdvancedStatefulSetType:
		c.Asts.OwnerReferences = references
	}
}

func (c *CacheWorkerSet) GetManagedFields() []metav1.ManagedFieldsEntry {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.ManagedFields
	case DaemonSetType:
		return c.Ds.ManagedFields
	case AdvancedStatefulSetType:
		return c.Asts.ManagedFields
	default:
		return nil
	}
}

func (c *CacheWorkerSet) SetManagedFields(managedFields []metav1.ManagedFieldsEntry) {
	switch c.WorkerType {
	case StatefulSetType:
		c.Sts.ManagedFields = managedFields
	case DaemonSetType:
		c.Ds.ManagedFields = managedFields
	case AdvancedStatefulSetType:
		c.Asts.ManagedFields = managedFields
	}
}

func (c *CacheWorkerSet) GetObjectKind() schema.ObjectKind {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.GetObjectKind()
	case DaemonSetType:
		return c.Ds.GetObjectKind()
	case AdvancedStatefulSetType:
		return c.Asts.GetObjectKind()
	default:
		return nil
	}
}

func (c *CacheWorkerSet) DeepCopyObject() runtime.Object {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts.DeepCopy()
	case DaemonSetType:
		return c.Ds.DeepCopy()
	case AdvancedStatefulSetType:
		return c.Asts.DeepCopy()
	default:
		return nil
	}
}

// NewCacheWorkerManagerClass creates a new CacheWorkerManagerClass
func NewCacheWorker(client client.Client) *CacheWorkerSet {
	return &CacheWorkerSet{
		client: client,
	}
}

//	type CacheWorkerSet interface {
//		ToStatefulSet() *appsv1.StatefulSet
//		ToDaemonSet() *appsv1.DaemonSet
//		ToAdvancedStatefulSet() *openkruise.StatefulSet
//	}
func (c *CacheWorkerSet) ToResource() runtime.Object {
	switch c.WorkerType {
	case StatefulSetType:
		return c.Sts
	case DaemonSetType:
		return c.Ds
	case AdvancedStatefulSetType:
		return c.Asts
	default:
		return nil
	}
}
func (c *CacheWorkerSet) ToStatefulSet() *appsv1.StatefulSet {
	return c.Sts
}
func (c *CacheWorkerSet) ToDaemonSet() *appsv1.DaemonSet {
	return c.Ds
}
func (c *CacheWorkerSet) ToAdvancedStatefulSet() *openkruise.StatefulSet {
	return c.Asts
}
func GetWorkerAsCacheWorkerSet(c client.Client, name string, namespace string, WorkerType string) (*CacheWorkerSet, error) {
	if WorkerType == string(StatefulSetType) || WorkerType == "" {
		Sts, err := kubeclient.GetStatefulSet(c, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get StatefulSet: %")
		}
		return &CacheWorkerSet{
			client:     c,
			WorkerType: StatefulSetType,
			Sts:        Sts,
		}, nil
	} else if WorkerType == string(DaemonSetType) {
		Ds, err := kubeclient.GetDaemonset(c, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get DaemonSet: %w", err)
		}
		return &CacheWorkerSet{
			client:     c,
			WorkerType: DaemonSetType,
			Ds:         Ds,
		}, nil
	} else if WorkerType == string(AdvancedStatefulSetType) {
		Asts, err := kubeclient.GetAdvancedStatefulSet(c, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get AdvancedStatefulSet: %w", err)
		}
		return &CacheWorkerSet{
			client:     c,
			WorkerType: AdvancedStatefulSetType,
			Asts:       Asts,
		}, nil
	}
	return nil, fmt.Errorf("unsupported WorkerType '%s': %w", WorkerType, fluiderrs.NewNotSupported(schema.GroupResource{
		Group:    "fluid",
		Resource: "CacheWorkerSet",
	}, WorkerType))
}

func StsToCacheWorkerSet(set *appsv1.StatefulSet) *CacheWorkerSet {
	return &CacheWorkerSet{
		WorkerType: StatefulSetType,
		Sts:        set,
	}
}
func DsToCacheWorkerSet(ds *appsv1.DaemonSet) *CacheWorkerSet {
	return &CacheWorkerSet{
		WorkerType: DaemonSetType,
		Ds:         ds,
	}
}
func AstsToCacheWorkerSet(asts *openkruise.StatefulSet) *CacheWorkerSet {
	return &CacheWorkerSet{
		WorkerType: AdvancedStatefulSetType,
		Asts:       asts,
	}
}

// ScaleStatefulSet scale the statefulset replicas
func ScaleAdvancedStatefulSet(client client.Client, name string, namespace string, replicas int32) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := kubeclient.GetAdvancedStatefulSet(client, name, namespace)
		if err != nil {
			return err
		}
		workersToUpdate := workers.DeepCopy()
		workersToUpdate.Spec.Replicas = &replicas
		if !reflect.DeepEqual(workers, workersToUpdate) {
			err = client.Update(context.TODO(), workersToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func GetWorkersAsAdvancedStatefulset(client client.Client, key types.NamespacedName) (workers *openkruise.StatefulSet, err error) {
	workers, err = kubeclient.GetAdvancedStatefulSet(client, key.Name, key.Namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			_, dsErr := kubeclient.GetDaemonset(client, key.Name, key.Namespace)
			// return workers, fluiderr.NewDeprecated()
			// find the daemonset successfully
			if dsErr == nil {
				return workers, fluiderrs.NewDeprecated(schema.GroupResource{
					Group:    appsv1.SchemeGroupVersion.Group,
					Resource: "daemonsets",
				}, key)
			}
		}
	}

	return
}
