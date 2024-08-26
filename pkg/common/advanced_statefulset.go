package common

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"strconv"
)

const (
	advancedStatefulSetEnabled = false
)

// ScaleConfig 用户输入的缩容配置
type ScaleConfig struct {
	WorkerType     WorkerType
	Replicas       *int32
	OfflineIndices []int // 仅在 AdvancedStatefulSet 中有效
}

// UnmarshalYAML implements the Unmarshaler interface for custom handling of YAML data.
// 从yaml中读取出offlineIndices
func (config *ScaleConfig) UnmarshalYAML(unmarshaler *runtime.Decoder, data []byte) error {
	if err := unmarshaler.Decode(data, &config.WorkerType, &config.Replicas); err != nil {
		return err
	}

	// Check if AdvancedStatefulSet feature is enabled
	if features, err := unmarshaler.DecodeAsMap(data); err == nil {
		if _, ok := features["features"]; ok {
			// Implement further logic to check if 'AdvancedStatefulSet' feature is enabled
			// This is a placeholder, you'd need to extend this logic based on your actual implementation
			if advancedStatefulSetEnabled {
				var offlineIndices []int
				if err := unmarshaler.Decode(&offlineIndices); err == nil {
					config.OfflineIndices = offlineIndices
				}
			}
		}
	}

	// If AdvancedStatefulSet is enabled but OfflineIndices field is not present, it should be an error according to your spec

	return nil
}

// Deserialize decodes YAML data into a generic object.
func Deserialize(data []byte) (*store.Store, error) {
	// Ensure the object is registered with the deserializer.
	gvr := schema.GroupVersionResource{Group: "data.fluid.io", Version: "v1alpha1", Resource: "alluxiuruntimes"}
	runtime.NewRuntime(
		runtime.APIGroupUnExportedObject{Group: &gvr.Group, Version: &gvr.Version, Kind: &schema.KnownKind{}, Plural: &gvr.Resource}).
		MustRegisterUnstructured(AlluxioRuntime{})

	// Deserialize the YAML.
	var object runtime.RawExtension
	if err := runtime.DecodeInto(&object, data, runtime.NewSerializer()); err != nil {
		return nil, err
	}

	// Deserialize the JSON.
	var store *store.Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	return store, nil
}

// Function to get workers by type.
func getWorkersByType(workers []alluxio.Worker, workerType string) []alluxio.Worker {
	var result []alluxio.Worker
	for _, worker := range workers {
		if typ, ok := worker.Annotations["workerType"]; ok && typ == workerType {
			result = append(result, worker)
		}
	}
	return result
}

// Function to process AdvancedStatefulSet workers.
func processAdvancedStatefulSetWorkers(workers []alluxio.Worker) {
	for _, worker := range workers {
		if workerType, ok := worker.Annotations["workerType"]; ok && workerType == "AdvancedStatefulSet" {
			if scaleDownStr, ok := worker.Annotations["scaleDown"]; ok && scaleDownStr == "true" {
				// Here, we assume that the pod index is stored in a field or annotation, for simplicity, let's use a placeholder.
				// Example: worker.Annotations["podIndex"] would be used to retrieve the pod index.
				if podIndexStr, ok := worker.Annotations["podIndex"]; ok {
					// Convert podIndexStr to int and add to scaleDown list
					podIndex, err := strconv.Atoi(podIndexStr)
					if err == nil {
						scaleDown = append(scaleDown, podIndex)
					}
				}
			}
		}
	}
}







//2--------------------2-------------------2-----2--------------------2-------------------2-----2--------------------2-------------------2-----

const (
	advancedStatefulSetEnabled = false
)

// ScaleConfig 用户输入的缩容配置
type ScaleConfig struct {
	Replicas       *int32
	OfflineIndices []int // 仅在 AdvancedStatefulSet 中有效
}

// AdvancedStatefulSetSpec 定义了 StatefulSet 的规范部分
type AdvancedStatefulSetSpec struct {
}

type AdvancedStatefulSetStatus struct {
}
type AdvancedStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              AdvancedStatefulSetSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            AdvancedStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// 第一种方式，直接从worker的Annotations中读取
// Function to get workers by type.
func getTypesByWorker(workers []alluxio.Worker, workerType string) []alluxio.Worker {
	var result []alluxio.Worker
	for _, worker := range workers {
		if typ, ok := worker.Annotations["workerType"]; ok && typ == workerType {
			result = append(result, worker)
		}
	}
	return result
}

// 该函数根据 scaleDown 列表中的 pod 索引为这些 worker 打上 scaleDown: true 的注解
func annotateWorkersForScaleDown(workers []alluxio.Worker, scaleDown []int) []alluxio.Worker {
	// 创建一个映射，用于快速检查需要打上注解的 pod 索引
	scaleDownSet := make(map[int]struct{})
	for _, idx := range scaleDown {
		scaleDownSet[idx] = struct{}{}
	}

	// 遍历所有 workers，并为需要打上注解的 worker 打上 scaleDown: true 的注解
	for i, worker := range workers {
		if _, ok := scaleDownSet[i]; ok {
			if worker.Annotations == nil {
				worker.Annotations = make(map[string]string)
			}
			worker.Annotations["scaleDown"] = "true"
			// 更新 worker 列表中的 worker
			workers[i] = worker
		}
	}

	return workers
}

// 通过worker的标签获取workerType为AdvancedStatefulSet并且要下线的节点
func processAdvancedStatefulSetWorkersAndGetScaleDownIndicesByWorkerAnno(workers []alluxio.Worker) ScaleConfig {
	var scaleDown []int
	for _, worker := range workers {
		if workerType, ok := worker.Annotations["workerType"]; ok && workerType == "AdvancedStatefulSet" {
			if scaleDownStr, ok := worker.Annotations["scaleDown"]; ok && scaleDownStr == "true" {
				if podIndexStr, ok := worker.Annotations["podIndex"]; ok {
					podIndex, err := strconv.Atoi(podIndexStr)
					if err == nil {
						scaleDown = append(scaleDown, podIndex)
					}
				}
			}
		}
	}
	return ScaleConfig{
		OfflineIndices: scaleDown,
	}
}

// 第二种方式，从statefulset中增加一个metadata字段中读取delete-slots ?
// apiVersion: apps.pingcap.com/v1alpha1
// kind: StatefulSet
// metadata:
// name: web
// annotations:
// delete-slots: '[1]'
// spec:
// selector:
// matchLabels:
// app: nginx
// serviceName: "nginx"
// replicas: 2
// template:
// metadata:
// labels:
// app: nginx
// spec:
// terminationGracePeriodSeconds: 10
// containers:
// - name: nginx
// image: k8s.gcr.io/nginx-slim:0.8
// ports:
// - containerPort: 80
// name: web
// revisionHistoryLimit: 10

