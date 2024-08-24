package common

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

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
func getWorkersByType(workers []Worker, workerType string) []Worker {
	var result []Worker
	for _, worker := range workers {
		if typ, ok := worker.Annotations["workerType"]; ok && typ == workerType {
			result = append(result, worker)
		}
	}
	return result
}

// Function to process AdvancedStatefulSet workers.
func processAdvancedStatefulSetWorkers(workers []Worker) {
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
