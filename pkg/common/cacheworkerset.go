package common

import (
	"fmt"
	"golang.org/x/net/context"
	"k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// NewResourceManager 创建一个新的 ResourceManager
func NewResourceManager(client client.Client, cache CacheWorkerSet) *ResourceManager {
	return &ResourceManager{client: client, cache: cache}
}

// GetWorkerType 从 StatefulSet 的注解中读取 workerType
func (rm *ResourceManager) GetWorkerType(ctx context.Context, runtimeName string, namespace string) (WorkerType, error) {
	// 获取 StatefulSet 对象
	statefulSet, err := rm.clientset.AppsV1().StatefulSets(namespace).Get(ctx, runtimeName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get StatefulSet %s in namespace %s: %v", runtimeName, namespace, err)
	}

	// 从注解中读取 workerType
	annotations := statefulSet.ObjectMeta.Annotations
	workerType, exists := annotations["worker-type"]
	if !exists {
		// 如果没有找到 worker-type 注解，返回默认值 StatefulSetType
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

// ProcessScaleConfig 处理用户输入的缩容配置并执行缩容操作
func (rm *ResourceManager) ProcessScaleConfig(ctx context.Context, runtimeName string, namespace string, config ScaleConfig) error {
	workerType, err := rm.GetWorkerType(ctx, runtimeName, namespace)
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
