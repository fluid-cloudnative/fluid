package common

import (
	"encoding/json"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CacheWorkerSet 接口定义了对 StatefulSet 和 DaemonSet 的操作
type CacheWorkerSet interface {
	// Update 更新 CacheWorkerSet 的定义
	Update() error

	// Delete 删除 CacheWorkerSet
	Delete() error

	// GetStatus 获取 CacheWorkerSet 的当前状态
	GetStatus() (interface{}, error)

	// Type 返回资源类型，例如 "StatefulSet" 或 "DaemonSet"
	Type() string
}

// StatefulSetCacheWorker 实现了 CacheWorkerSet 接口，专门用于操作 StatefulSet
type StatefulSetCacheWorker struct {
	Set *v1.StatefulSet
}

func (s *StatefulSetCacheWorker) Update() error {
	// 实现 StatefulSet 更新逻辑
	return nil
}

func (s *StatefulSetCacheWorker) Delete() error {
	// 实现 StatefulSet 删除逻辑
	return nil
}

func (s *StatefulSetCacheWorker) GetStatus() (interface{}, error) {
	// 实现获取 StatefulSet 状态逻辑
	return s.Set.Status, nil
}

func (s *StatefulSetCacheWorker) Type() string {
	return "StatefulSet"
}

// DaemonSetCacheWorker 实现了 CacheWorkerSet 接口，专门用于操作 DaemonSet
type DaemonSetCacheWorker struct {
	Set *v1.DaemonSet
}

func (d *DaemonSetCacheWorker) Update() error {
	// 实现 DaemonSet 更新逻辑
	return nil
}

func (d *DaemonSetCacheWorker) Delete() error {
	// 实现 DaemonSet 删除逻辑
	return nil
}

func (d *DaemonSetCacheWorker) GetStatus() (interface{}, error) {
	// 实现获取 DaemonSet 状态逻辑
	return d.Set.Status, nil
}

func (d *DaemonSetCacheWorker) Type() string {
	return "DaemonSet"
}

// StatefulSetCacheWorker 实现了 CacheWorkerSet 接口，专门用于操作 StatefulSet
type AdvancedStatefulSetCacheWorker struct {
	Set *AdvancedStatefulSet
}

func (as *AdvancedStatefulSetCacheWorker) Update() error {
	// 实现 StatefulSet 更新逻辑
	return nil
}

func (as *AdvancedStatefulSetCacheWorker) Delete() error {
	// 实现 StatefulSet 删除逻辑
	return nil
}

func (as *AdvancedStatefulSetCacheWorker) GetStatus() (interface{}, error) {
	// 实现获取 StatefulSet 状态逻辑
	return as.Set.Status, nil
}

func (as *AdvancedStatefulSetCacheWorker) Type() string {
	return "AdvancedStatefulSet"
}

// NewCacheWorkerSet 根据提供的资源类型创建对应的 CacheWorkerSet 实例
func NewCacheWorkerSet(kind string, set interface{}) (CacheWorkerSet, error) {
	switch kind {
	case "StatefulSet":
		if set, ok := set.(*v1.StatefulSet); ok {
			return &StatefulSetCacheWorker{Set: set}, nil
		}
	case "DaemonSet":
		if set, ok := set.(*v1.DaemonSet); ok {
			return &DaemonSetCacheWorker{Set: set}, nil
		}
	case "AdvancedStatefulSet":
		if set, ok := set.(*AdvancedStatefulSet); ok {
			return &AdvancedStatefulSetCacheWorker{Set: set}, nil
		}
	}
	return nil, fmt.Errorf("unsupported kind %s", kind)
}

// WorkerType 定义 worker 管理类型
type WorkerType string

const workerSetTypeKey string = "data.fluid.io/workerType"

const (
	// 默认的 StatefulSet 类型
	StatefulSetType WorkerType = "StatefulSet"
	// DaemonSet 类型
	DaemonSetType WorkerType = "DaemonSet"
	// 高级 StatefulSet 类型
	AdvancedStatefulSetType WorkerType = "AdvancedStatefulSet"

	DefaultWorkerSetType WorkerType = StatefulSetType
)

// ResourceManager 负责管理不同类型的 worker 并执行缩容操作
type ResourceManager struct {
	client client.Client
	cache  CacheWorkerSet
}

func GetAnnotationKeyOfWorkerType(accessor metav1.ObjectMetaAccessor) (ret string) {
	annotations := accessor.GetObjectMeta().GetAnnotations()
	if annotations == nil {
		return
	}
	m := annotations[workerSetTypeKey]
	if m == "" {
		return
	}
	if err := json.Unmarshal([]byte(m), &ret); err != nil {
		log := ctrl.Log.WithName("base")
		log.V(5).Error(err, "failed to unmarshal workerType from annotations", "data.fluid.io/workerType", m)
	}
	return
}

// 处理 workerType 的函数
func processWorkerType(workerType *WorkerType) {
	// 如果 workerType 为空，设置为默认值
	if *workerType == "" {
		*workerType = DefaultWorkerSetType
	} else if *workerType == StatefulSetType || *workerType == DaemonSetType || *workerType == AdvancedStatefulSetType {
		return
	} else {
		// 否则设置为错误类型
		fmt.Println("Error: invalid worker type")
	}
}

// GetWorkerType 从 StatefulSet 的注解中读取 workerType
// 以alluxioRuntime为例
func (rm *ResourceManager) GetSpecifiedWorkerSet(namespace []types.NamespacedName) (CacheWorkerSet, error) {
	// 获取 StatefulSet 对象

	//通过dataset获取runtime类型
	//dataset, err := utils.GetDataset(rm.client, namespace[0].Name, namespace[0].Namespace)
	//if len(dataset.Status.Runtimes) != 0 {
	//	runtimeType := dataset.Status.Runtimes[0].Type
	//}
	alluxioRuntime, err := utils.GetAlluxioRuntime(rm.client, namespace[0].Name, namespace[0].Namespace)
	if err != nil {
		return nil, err
	}
	// 使用 GetAnnotationKeyOfWorkerType 函数获取annotation
	var workerType string
	workerType = GetAnnotationKeyOfWorkerType(alluxioRuntime) // 这里调用获取 workerType 的函数

	processWorkerType((*WorkerType)(&workerType))
	var set interface{}
	SpecifiedWorkerSet, _ := NewCacheWorkerSet(workerType, set)
	return SpecifiedWorkerSet, nil
}
