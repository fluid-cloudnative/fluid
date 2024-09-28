package cacheworkerset

import (
	"context"
	"fmt"
	v12 "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/client/v1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkerType defines the type of worker resource
type WorkerType string
type StatefulSetConditionType string

const (
	StatefulSetType         WorkerType = "statefulset"
	AdvancedStatefulSetType WorkerType = "advanced_statefulset"
	DaemonSetType           WorkerType = "daemonset"
)

// CacheWorkerManagerClass defines the manager class
type CacheWorkerManagerClass struct {
	client client.Client
}

// NewCacheWorkerManagerClass creates a new CacheWorkerManagerClass
func NewCacheWorkerManagerClass(client client.Client) *CacheWorkerManagerClass {
	return &CacheWorkerManagerClass{
		client: client,
	}
}

// CacheWorkerSet defines the worker set
type CacheWorkerSet struct {
	Spec   CacheWorkerSetSpec
	Status CacheWorkerSetStatus
}

// CacheWorkerSetSpec defines the spec for CacheWorkerSet
type CacheWorkerSetSpec struct {
	Type WorkerType
	// Other spec fields
}
type CacheWorkerSetCondition struct {
	// Type of statefulset condition.
	Type StatefulSetConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// CacheWorkerSetStatus defines the status for CacheWorkerSet
type CacheWorkerSetStatus struct {
	Replicas           int
	WorkerType         WorkerType
	ObservedGeneration int64
	ReadyReplicas      int32
	CurrentReplicas    int32
	UpdatedReplicas    int32
	CurrentRevision    string
	UpdateRevision     string
	CollisionCount     *int32
	Conditions         []CacheWorkerSetCondition
	// Other status fields
}

// GetWorker gets the worker based on the provided type
func GetWorker(client client.Client, key types.NamespacedName) (*CacheWorkerSet, error) {
	var workerStatus WorkerSetStatus
	var err error

	// First try to get as StatefulSet
	workerStatus, err = getStatefulSetStatus(client, key)
	if err == nil {
		return &CacheWorkerSet{
			Spec: CacheWorkerSetSpec{
				Type: StatefulSetType,
			},
			Status: CacheWorkerSetStatus{
				Replicas:   workerStatus.GetReplicas(),
				WorkerType: StatefulSetType,
				// Other status fields
			},
		}, nil
	}

	// If StatefulSet is not found, try to get as AdvancedStatefulSet
	workerStatus, err = getAdvancedStatefulSetStatus(client, key)
	if err == nil {
		return &CacheWorkerSet{
			Spec: CacheWorkerSetSpec{
				Type: AdvancedStatefulSetType,
			},
			Status: CacheWorkerSetStatus{
				Replicas:   workerStatus.GetReplicas(),
				WorkerType: AdvancedStatefulSetType,
				// Other status fields
			},
		}, nil
	}

	// If AdvancedStatefulSet is not found, try to get as DaemonSet
	workerStatus, err = getDaemonSetStatus(client, key)
	if err == nil {
		return &CacheWorkerSet{
			Spec: CacheWorkerSetSpec{
				Type: DaemonSetType,
			},
			Status: CacheWorkerSetStatus{
				Replicas:   workerStatus.GetReplicas(),
				WorkerType: DaemonSetType,
				// Other status fields
			},
		}, nil
	}
	// If none of the resources are found, return an error
	return nil, fmt.Errorf("worker not found for key: %s", key.Name)
}

func getStatefulSetStatus(client client.Client, key types.NamespacedName) (WorkerSetStatus, error) {
	statefulSet := &appsv1.StatefulSet{}
	statefulSet, err := kubeclient.GetStatefulSet(client, key.Name, key.Namespace)
	//err := client.Get(context.TODO(), key, statefulSet)
	if err != nil {
		return nil, err
	}
	return &StatefulSetWorkerStatus{StatefulSet: statefulSet}, nil
}

func getAdvancedStatefulSetStatus(client client.Client, key types.NamespacedName) (WorkerSetStatus, error) {
	advancedStatefulSet := &v12.AdvancedStatefulSet{}
	err := client.Get(context.TODO(), key, advancedStatefulSet)
	if err != nil {
		return nil, err
	}
	return &AdvancedStatefulSetWorkerStatus{AdvancedStatefulSet: advancedStatefulSet}, nil
}

func getDaemonSetStatus(client client.Client, key types.NamespacedName) (WorkerSetStatus, error) {
	daemonSet := &appsv1.DaemonSet{}
	daemonSet, err := kubeclient.GetDaemonset(client, key.Name, key.Namespace)
	//err := client.Get(context.TODO(), key, daemonSet)
	if err != nil {
		return nil, err
	}
	return &DaemonSetWorkerStatus{DaemonSet: daemonSet}, nil
}

// WorkerSetStatus defines the interface for worker set status
type WorkerSetStatus interface {
	GetReplicas() int
	GetSpec() interface{}
	// Other status-related methods
}

// DaemonSetWorkerStatus implements WorkerSetStatus for DaemonSets
type DaemonSetWorkerStatus struct {
	DaemonSet *appsv1.DaemonSet
}

// StatefulSetWorkerStatus implements WorkerSetStatus for StatefulSets
type StatefulSetWorkerStatus struct {
	StatefulSet *appsv1.StatefulSet
}

// AdvancedStatefulSetWorkerStatus implements WorkerSetStatus for AdvancedStatefulSets
type AdvancedStatefulSetWorkerStatus struct {
	AdvancedStatefulSet *v12.AdvancedStatefulSet
}

func (s *AdvancedStatefulSetWorkerStatus) GetReplicas() int {
	return int(*s.AdvancedStatefulSet.Spec.Replicas)
}

func (s *AdvancedStatefulSetWorkerStatus) GetSpec() interface{} {
	return s.AdvancedStatefulSet.Spec
}

func (s *StatefulSetWorkerStatus) GetReplicas() int {
	return int(*s.StatefulSet.Spec.Replicas)
}

func (s *StatefulSetWorkerStatus) GetSpec() interface{} {
	return s.StatefulSet.Spec
}

func (d *DaemonSetWorkerStatus) GetReplicas() int {
	return int(d.DaemonSet.Status.DesiredNumberScheduled)
}

func (d *DaemonSetWorkerStatus) GetSpec() interface{} {
	return d.DaemonSet.Spec
}
