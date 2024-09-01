package cacheworker

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AdvancedStatefulSetSpec struct {
	scaleDownIndices []int
	Replicas         int
}

// StatefulSetStatus 实现 WorkerSetStatus 接口
type AdvancedStatefulSetStatus struct {
	scaleDownIndices []int
	Replicas         int
	as               AdvancedStatefulSet
}

func (a AdvancedStatefulSetStatus) GetReplicas() int {
	//TODO implement me
	panic("implement me")
}

func (a AdvancedStatefulSetStatus) GetSpec() interface{} {
	//TODO implement me
	panic("implement me")
}

type AdvancedStatefulSet struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// The desired behavior of this daemon set.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	//spec AdvancedStatefulSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// The current status of this daemon set. This data may be
	// out of date by some window of time.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	//Status AdvancedStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (a AdvancedStatefulSet) DeepCopyObject() runtime.Object {
	//TODO implement me
	panic("implement me")
}
