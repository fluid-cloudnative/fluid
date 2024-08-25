package common

import (
	"encoding/json"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// 通用接口，处理所有资源的通用操作
type ResourceHandler interface {
	Scale(client client.Client, name string, namespace string, replicas int32) error
	GetPodsForResource(c client.Client, name string, namespace string) ([]v1.Pod, error)
	GetPhaseFromResource(replicas int32, resource interface{}) (string, error)
	GetUnavailablePodsForResource(c client.Client, resource interface{}, selector labels.Selector) ([]*v1.Pod, error)
	GetUnavailablePodNamesForResource(c client.Client, resource interface{}, selector labels.Selector) ([]types.NamespacedName, error)
}

// 处理 DaemonSet 的接口
type DaemonSetHandler interface {
	ResourceHandler
	GetDaemonSet(c client.Client, name string, namespace string) (*v1.DaemonSet, error)
}

// 处理 StatefulSet 的接口
type StatefulSetHandler interface {
	ResourceHandler
	GetStatefulSet(c client.Client, name string, namespace string) (*v1.StatefulSet, error)
	GetPodsForStatefulSet(c client.Client, sts *v1.StatefulSet, selector labels.Selector) ([]v1.Pod, error)
	IsMemberOfStatefulSet(sts *v1.StatefulSet, pod *v1.Pod) bool
	GetParentNameAndOrdinal(pod *v1.Pod) (string, int)
	GetParentName(pod *v1.Pod) string
}

// 处理 AdvancedStatefulSet 的接口（假设 AdvancedStatefulSet 是 StatefulSet 的扩展）
type AdvancedStatefulSetHandler interface {
	StatefulSetHandler
	// 如果有额外的特性，可以在这里定义
}

// 组合接口，支持所有资源
type CacheWorkerSet interface {
	DaemonSetHandler
	StatefulSetHandler
	AdvancedStatefulSetHandler
}

// WorkerType 定义 worker 管理类型
type WorkerType string

const (
	// 默认的 StatefulSet 类型
	StatefulSetType WorkerType = "StatefulSet"
	// DaemonSet 类型
	DaemonSetType WorkerType = "DaemonSet"
	// 高级 StatefulSet 类型
	AdvancedStatefulSetType WorkerType = "AdvancedStatefulSet"
)

// ScaleConfig 用户输入的缩容配置
type ScaleConfig struct {
	WorkerType     WorkerType
	Replicas       *int32
	OfflineIndices []int // 仅在 AdvancedStatefulSet 中有效
}

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
	m := annotations["data.fluid.io/workerType"]
	if m == "" {
		return
	}
	if err := json.Unmarshal([]byte(m), &ret); err != nil {
		log := ctrl.Log.WithName("base")
		log.V(5).Error(err, "failed to unmarshal workerType from annotations", "data.fluid.io/workerType", m)
	}
	return
}

// GetWorkerType 从 StatefulSet 的注解中读取 workerType
// 以alluxioRuntime为例
func (rm *ResourceManager) GetWorkerType(ctx context.Context, namespace []types.NamespacedName) (WorkerType, error) {
	// 获取 StatefulSet 对象

	dataset, err := utils.GetDataset(rm.client, namespace[0].Name, namespace[0].Namespace)
	var runtimeType string
	if len(dataset.Status.Runtimes) != 0 {
		runtimeType = dataset.Status.Runtimes[0].Type
	}
	alluxioRuntime, err := utils.GetAlluxioRuntime(rm.client, namespace[0].Name, namespace[0].Namespace)
	if err != nil {
		return "", err
	}
	// 使用 GetAnnotationKeyOfWorkerType 函数获取annotation
	var workerType string
	workerType = GetAnnotationKeyOfWorkerType(alluxioRuntime) // 这里调用获取 workerType 的函数

	// 处理从注解中获取的 workerType
	if workerType == "" {
		// 如果没有找到 metadata，返回默认值 StatefulSetType
		return StatefulSetType, nil
	}

	// 返回 workerType
	switch workerType {
	case "StatefulSet":
		return StatefulSetType, nil
	// 可以添加其他 worker 类型的处理逻辑
	default:
		return "", fmt.Errorf("unknown worker-type value: %s", workerType)
	}
}

// NewResourceManager 创建一个新的 ResourceManager
func NewResourceManager(client client.Client, cache CacheWorkerSet) *ResourceManager {
	return &ResourceManager{client: client, cache: cache}
}

func (rm *ResourceManager) ProcessCacheWorkers(ctx context.Context, runtimeName string, namespace string, config ScaleConfig) error {
	return nil
}

// ProcessScaleConfig 处理用户输入的缩容配置并执行缩容操作
func (rm *ResourceManager) ProcessScaleConfig(ctx context.Context, runtimeName string, Namespace []types.NamespacedName, config ScaleConfig) error {
	var namespace string
	namespace = Namespace[0].Namespace
	workerType, err := rm.GetWorkerType(ctx, Namespace)
	if err != nil {
		return fmt.Errorf("failed to get worker type: %w", err)
	}

	var errMsg string
	switch workerType {
	case StatefulSetType:
		if config.Replicas == nil {
			return fmt.Errorf("replicas must be provided for StatefulSet")
		}
		statefulSet, err := rm.cache.GetStatefulSet(rm.client, runtimeName, namespace)
		if err != nil {
			return fmt.Errorf("failed to get StatefulSet: %w", err)
		}
		if err := rm.cache.Scale(rm.client, runtimeName, namespace, *config.Replicas); err != nil {
			return fmt.Errorf("failed to scale StatefulSet: %w", err)
		}
	// 添加对应的 DaemonSet 和 AdvancedStatefulSet 处理
	case DaemonSetType:
		if config.Replicas == nil {
			return fmt.Errorf("replicas must be provided for DaemonSet")
		}
		daemonSet, err := rm.cache.GetDaemonSet(rm.client, runtimeName, namespace)
		if err != nil {
			return fmt.Errorf("failed to get DaemonSet: %w", err)
		}
		if err := rm.cache.Scale(rm.client, runtimeName, namespace, *config.Replicas); err != nil {
			return fmt.Errorf("failed to scale DaemonSet: %w", err)
		}
	case AdvancedStatefulSetType:
		if len(config.OfflineIndices) == 0 {
			return fmt.Errorf("offlineIndices must be provided for AdvancedStatefulSet")
		}
		statefulSet, err := rm.cache.GetStatefulSet(rm.client, runtimeName, namespace)
		if err != nil {
			return fmt.Errorf("failed to get StatefulSet: %w", err)
		}
		for _, index := range config.OfflineIndices {
			podName := fmt.Sprintf("%s-%d", statefulSet.Name, index)
			pod := &v1.Pod{}
			err = rm.client.Get(ctx, types.NamespacedName{Name: podName, Namespace: namespace}, pod)
			if err != nil {
				return fmt.Errorf("failed to get Pod %s: %w", podName, err)
			}
			if err := rm.client.Delete(ctx, pod); err != nil {
				return fmt.Errorf("failed to delete Pod %s: %w", podName, err)
			}
		}
	default:
		errMsg = "unsupported worker type"
	}

	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}

	return nil
}
