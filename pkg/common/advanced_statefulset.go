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

// Global variable to store scale down pod indices.
var scaleDown []int

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

// Function to process AdvancedStatefulSet workers.
func processAdvancedStatefulSetWorkersByWorkerAnno(workers []alluxio.Worker) {
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

func main() {
	// Sample workers data (usually this would come from a source like Kubernetes API)
	workers := []alluxio.Worker{
		{Annotations: map[string]string{"workerType": "StatefulSet"}},
		{Annotations: map[string]string{"workerType": "AdvancedStatefulSet", "scaleDown": "true", "podIndex": "1"}},
		{Annotations: map[string]string{"workerType": "AdvancedStatefulSet", "scaleDown": "false", "podIndex": "2"}},
		{Annotations: map[string]string{"workerType": "DaemonSet"}},
	}

	// Get workers by type
	statefulSetWorkers := getTypesByWorker(workers, "StatefulSet")
	daemonSetWorkers := getTypesByWorker(workers, "DaemonSet")
	advancedStatefulSetWorkers := getTypesByWorker(workers, "AdvancedStatefulSet")

	fmt.Printf("StatefulSet Workers: %+v\n", statefulSetWorkers)
	fmt.Printf("DaemonSet Workers: %+v\n", daemonSetWorkers)

	// Process AdvancedStatefulSet Workers
	processAdvancedStatefulSetWorkersByWorkerAnno(advancedStatefulSetWorkers)

	fmt.Printf("Scale Down List: %v\n", scaleDown)
}

