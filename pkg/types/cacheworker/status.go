package cacheworker

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkerType string

const (
	StatefulSetType         WorkerType = "statefulset"
	AdvancedStatefulSetType WorkerType = "advanced_statefulset"
	DaemonSetType           WorkerType = "daemonset"
)

type CacheWorkerManagerClass struct {
	client client.Client
}

func NewCacheWorkerManagerClass(client client.Client) *CacheWorkerManagerClass {
	return &CacheWorkerManagerClass{
		client: client,
	}
}

func (m *CacheWorkerManagerClass) GetWorker(ctx context.Context, key types.NamespacedName, workerType WorkerType) (WorkerSetStatus, error) {
	// 默认为 StatefulSetType
	if workerType == "" {
		workerType = StatefulSetType
	}

	switch workerType {
	case StatefulSetType:
		return m.getStatefulSetStatus(ctx, key)
	case AdvancedStatefulSetType:
		// 需要实现获取高级 StatefulSet 的逻辑
		return m.getAdvancedStatefulSetStatus(ctx, key)
	case DaemonSetType:
		return m.getDaemonSetStatus(ctx, key)
	default:
		return nil, fmt.Errorf("unknown worker type: %s", workerType)
	}
}

func (m *CacheWorkerManagerClass) getStatefulSetStatus(ctx context.Context, key types.NamespacedName) (WorkerSetStatus, error) {
	statefulSet, err := m.getStatefulSet(ctx, key)
	if err != nil {
		return nil, err
	}
	return &StatefulSetStatus{StatefulSet: statefulSet}, nil
}

func (m *CacheWorkerManagerClass) getAdvancedStatefulSetStatus(ctx context.Context, key types.NamespacedName) (WorkerSetStatus, error) {
	advancedStatefulSet, err := m.getAdvancedStatefulSet(ctx, key)
	if err != nil {
		return nil, err
	}
	return &AdvancedStatefulSetStatus{as: *advancedStatefulSet}, nil
}

func (m *CacheWorkerManagerClass) getDaemonSetStatus(ctx context.Context, key types.NamespacedName) (WorkerSetStatus, error) {
	daemonSet, err := m.getDaemonSet(ctx, key)
	if err != nil {
		return nil, err
	}
	return &DaemonSetStatus{DaemonSet: daemonSet}, nil
}

func (m *CacheWorkerManagerClass) getStatefulSet(ctx context.Context, key types.NamespacedName) (*appsv1.StatefulSet, error) {
	statefulSet := &appsv1.StatefulSet{}
	err := m.client.Get(ctx, key, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("statefulset %s not found", key.Name)
		}
		return nil, err
	}
	return statefulSet, nil
}

func (m *CacheWorkerManagerClass) getDaemonSet(ctx context.Context, key types.NamespacedName) (*appsv1.DaemonSet, error) {
	daemonSet := &appsv1.DaemonSet{}
	err := m.client.Get(ctx, key, daemonSet)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("daemonset %s not found", key.Name)
		}
		return nil, err
	}
	return daemonSet, nil
}
func (m *CacheWorkerManagerClass) getAdvancedStatefulSet(ctx context.Context, key types.NamespacedName) (*AdvancedStatefulSet, error) {
	asSet := &AdvancedStatefulSet{}
	err := m.client.Get(ctx, key, asSet)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("statefulset %s not found", key.Name)
		}
		return nil, err
	}
	return asSet, nil
}

// WorkerSetStatus 定义了不同 WorkerManager 的状态接口
type WorkerSetStatus interface {
	GetReplicas() int
	GetSpec() interface{}
	// 其他状态相关的接口方法

}

// StatefulSetStatus 实现 WorkerSetStatus 接口
type StatefulSetStatus struct {
	StatefulSet *appsv1.StatefulSet
}

func (s *StatefulSetStatus) GetReplicas() int {
	return int(*s.StatefulSet.Spec.Replicas)
}

func (s *StatefulSetStatus) GetSpec() interface{} {
	return s.StatefulSet.Spec
}

// DaemonSetStatus 实现 WorkerSetStatus 接口
type DaemonSetStatus struct {
	DaemonSet *appsv1.DaemonSet
}

func (d *DaemonSetStatus) GetReplicas() int {
	return int(d.DaemonSet.Status.DesiredNumberScheduled)
}

func (d *DaemonSetStatus) GetSpec() interface{} {
	return d.DaemonSet.Spec
}
