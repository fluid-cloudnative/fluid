package common

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type CacheWorkerSet interface {
	Get(id string) (interface{}, error)
	Update(obj interface{}) error
	Delete(id string) error
}

type StatefulSetCacheWorker struct {
	client kubernetes.Interface
}

func NewStatefulSetCacheWorker(client kubernetes.Interface) *StatefulSetCacheWorker {
	return &StatefulSetCacheWorker{client: client}
}

func (c *StatefulSetCacheWorker) Get(id string) (interface{}, error) {
	set, err := c.client.AppsV1().StatefulSets("default").Get(context.TODO(), id, metav1.GetOptions{})
	return set, err
}

func (c *StatefulSetCacheWorker) Update(obj interface{}) error {
	set := obj.(*appsv1.StatefulSet)
	_, err := c.client.AppsV1().StatefulSets(set.Namespace).Update(context.TODO(), set, metav1.UpdateOptions{})
	return err
}

func (c *StatefulSetCacheWorker) Delete(id string) error {
	err := c.client.AppsV1().StatefulSets("default").Delete(context.TODO(), id, metav1.DeleteOptions{})
	return err
}

type DaemonSetCacheWorker struct {
	client kubernetes.Interface
}

func NewDaemonSetCacheWorker(client kubernetes.Interface) *DaemonSetCacheWorker {
	return &DaemonSetCacheWorker{client: client}
}

func (c *DaemonSetCacheWorker) Get(id string) (interface{}, error) {
	set, err := c.client.AppsV1().DaemonSets("default").Get(context.TODO(), id, metav1.GetOptions{})
	return set, err
}

func (c *DaemonSetCacheWorker) Update(obj interface{}) error {
	set := obj.(*appsv1.DaemonSet)
	_, err := c.client.AppsV1().DaemonSets(set.Namespace).Update(context.TODO(), set, metav1.UpdateOptions{})
	return err
}

func (c *DaemonSetCacheWorker) Delete(id string) error {
	err := c.client.AppsV1().DaemonSets("default").Delete(context.TODO(), id, metav1.DeleteOptions{})
	return err
}

type AdvancedStatefulSetCacheWorker struct {
	client dynamic.Interface
}

type AdvancedStatefulSetSpec struct {
	Replicas         int32                  `json:"replicas"`
	Template         corev1.PodTemplateSpec `json:"template"`
	ScaleDownIndices []int                  `json:"scaleDownIndices,omitempty"` // 新增字段
}

// AdvancedStatefulSetStatus 描述了AdvancedStatefulSet的状态
type AdvancedStatefulSetStatus struct {
	// 添加任何你想要追踪的状态字段
	CurrentReplicas int32 `json:"currentReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AdvancedStatefulSet 是自定义资源的定义
type AdvancedStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdvancedStatefulSetSpec   `json:"spec,omitempty"`
	Status AdvancedStatefulSetStatus `json:"status,omitempty"`
}

// SchemeGroupVersion 是此API版本的schema.GroupVersion
var SchemeGroupVersion = schema.GroupVersion{Group: "yourgroup.example.com", Version: "v1"}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
func NewAdvancedStatefulSetCacheWorker(client dynamic.Interface) *AdvancedStatefulSetCacheWorker {
	return &AdvancedStatefulSetCacheWorker{client: client}
}

func (c *AdvancedStatefulSetCacheWorker) Get(id string) (interface{}, error) {
	set, err := c.client.Resource(your_custom_api.Resource()).Namespace("default").Get(context.TODO(), id, metav1.GetOptions{})
	return set, err
}

func (c *AdvancedStatefulSetCacheWorker) Update(obj interface{}) error {
	set := obj.(runtime.Object)
	_, err := c.client.Resource(your_custom_api.Resource()).Namespace("default").Update(context.TODO(), set, metav1.UpdateOptions{})
	return err
}

func (c *AdvancedStatefulSetCacheWorker) Delete(id string) error {
	err := c.client.Resource(your_custom_api.Resource()).Namespace("default").Delete(context.TODO(), id, metav1.DeleteOptions{})
	return err
}

func (c *AdvancedStatefulSetCacheWorker) ScaleDownPods(as *your_custom_api.AdvancedStatefulSet) error {
	if as.Spec.ScaleDownIndices == nil || len(as.Spec.ScaleDownIndices) == 0 {
		return fmt.Errorf("no scale down indices provided")
	}

	// 降低副本数
	currentReplicas := as.Spec.Replicas
	for _, index := range as.Spec.ScaleDownIndices {
		if index < int(currentReplicas) {
			currentReplicas--
		}
	}
	as.Spec.Replicas = currentReplicas

	// 更新AdvancedStatefulSet
	_, err := c.client.Resource(your_custom_api.Resource()).Namespace("default").Update(context.TODO(), as, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	// 删除指定序号的Pods
	for _, index := range as.Spec.ScaleDownIndices {
		podName := fmt.Sprintf("%s-%d", as.Name, index)
		err := c.client.CoreV1().Pods("default").Delete(context.TODO(), podName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// 解析scale-down-indices Annotation
func parseScaleDownIndices(annotations map[string]string) ([]int, error) {
	indicesStr, ok := annotations["scale-down-indices"]
	if !ok {
		return nil, fmt.Errorf("missing scale-down-indices annotation")
	}

	var indices []int
	if err := json.Unmarshal([]byte(indicesStr), &indices); err != nil {
		return nil, fmt.Errorf("failed to parse scale-down-indices: %w", err)
	}
	return indices, nil
}

// 示例：从StatefulSet中读取并解析scaleDownIndices
func handleStatefulSet(ss *appsv1.StatefulSet) error {
	indices, err := parseScaleDownIndices(ss.ObjectMeta.Annotations)
	if err != nil {
		return err
	}

	// 使用indices进行相应的操作...
	// 例如，你可以在这里调用ScaleDownPodsByIndices(indices)
	return nil
}

type AdvancedStatefulSetManager struct {
	StatefulSet *appsv1.StatefulSet
	Client      kubernetes.Interface
}

func NewAdvancedStatefulSetManager(client kubernetes.Interface, ss *appsv1.StatefulSet) *AdvancedStatefulSetManager {
	return &AdvancedStatefulSetManager{
		StatefulSet: ss,
		Client:      client,
	}
}
func (asm *AdvancedStatefulSetManager) Get() (*appsv1.StatefulSet, error) {
	// 实现获取StatefulSet的逻辑
	return asm.Client.AppsV1().StatefulSets(asm.StatefulSet.Namespace).Get(context.TODO(), asm.StatefulSet.Name, metav1.GetOptions{})
}

func (asm *AdvancedStatefulSetManager) Update() error {
	// 实现更新StatefulSet的逻辑
	_, err := asm.Client.AppsV1().StatefulSets(asm.StatefulSet.Namespace).Update(context.TODO(), asm.StatefulSet, metav1.UpdateOptions{})
	return err
}

func (asm *AdvancedStatefulSetManager) Delete() error {
	// 实现删除StatefulSet的逻辑
	err := asm.Client.AppsV1().StatefulSets(asm.StatefulSet.Namespace).Delete(context.TODO(), asm.StatefulSet.Name, metav1.DeleteOptions{})
	return err
}
func (asm *AdvancedStatefulSetManager) ScaleDownByIndices(indices []int) error {
	// 从Annotations中解析scaleDownIndices
	annotations := asm.StatefulSet.ObjectMeta.Annotations
	if indices == nil {
		indices, _ = parseScaleDownIndices(annotations)
	}

	// 减少副本数
	currentReplicas := asm.StatefulSet.Spec.Replicas
	for _, index := range indices {
		if index < int(*currentReplicas) {
			*currentReplicas = *currentReplicas - 1
		}
	}

	// 更新StatefulSet
	asm.StatefulSet.Spec.Replicas = currentReplicas
	err := asm.Update()
	if err != nil {
		return err
	}

	// 删除指定序号的Pods
	for _, index := range indices {
		podName := fmt.Sprintf("%s-%d", asm.StatefulSet.Name, index)
		err := asm.Client.CoreV1().Pods(asm.StatefulSet.Namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

type AdvancedStatefulSetManager struct {
	StatefulSet *appsv1.StatefulSet
	Client      kubernetes.Interface
}

func NewAdvancedStatefulSetManager(client kubernetes.Interface, ss *appsv1.StatefulSet) *AdvancedStatefulSetManager {
	return &AdvancedStatefulSetManager{
		StatefulSet: ss,
		Client:      client,
	}
}

// 实现CacheWorkerSet接口的Get方法
func (asm *AdvancedStatefulSetManager) Get(id string) (interface{}, error) {
	// 从Kubernetes获取StatefulSet
	ss, err := asm.Client.AppsV1().StatefulSets(asm.StatefulSet.Namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ss, nil
}

// 实现CacheWorkerSet接口的Update方法
func (asm *AdvancedStatefulSetManager) Update(obj interface{}) error {
	ss := obj.(*appsv1.StatefulSet)
	_, err := asm.Client.AppsV1().StatefulSets(ss.Namespace).Update(context.TODO(), ss, metav1.UpdateOptions{})
	return err
}

// 实现CacheWorkerSet接口的Delete方法
func (asm *AdvancedStatefulSetManager) Delete(id string) error {
	err := asm.Client.AppsV1().StatefulSets(asm.StatefulSet.Namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	return err
}

// 实现CacheWorkerSet接口的ScaleDownByIndices方法
func (asm *AdvancedStatefulSetManager) ScaleDownByIndices(id string, indices []int) error {
	// 从Annotations中解析scaleDownIndices
	annotations := asm.StatefulSet.ObjectMeta.Annotations
	parsedIndices, err := parseScaleDownIndices(annotations)
	if err != nil {
		return err
	}

	// 合并用户传入的indices和从Annotation中解析出的indices
	allIndices := append(parsedIndices, indices...)
	uniqueIndices := unique(allIndices)

	// 执行缩容操作
	err = asm.scaleDown(uniqueIndices)
	if err != nil {
		return err
	}

	return nil
}

// 辅助函数：去重
func unique(slice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// 辅助函数：执行缩容操作
func (asm *AdvancedStatefulSetManager) scaleDown(indices []int) error {
	// 减少副本数
	currentReplicas := asm.StatefulSet.Spec.Replicas
	for _, index := range indices {
		if index < int(*currentReplicas) {
			*currentReplicas = *currentReplicas - 1
		}
	}

	// 更新StatefulSet
	asm.StatefulSet.Spec.Replicas = currentReplicas
	err := asm.Update()
	if err != nil {
		return err
	}

	// 删除指定序号的Pods
	for _, index := range indices {
		podName := fmt.Sprintf("%s-%d", asm.StatefulSet.Name, index)
		err := asm.Client.CoreV1().Pods(asm.StatefulSet.Namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
