## How to create and delete fluid resources with Go client

The following is an example of creating and deleting Dataset and AlluxioRuntime, showing how to create and delete fluid resources by using the Go client.

### The yaml file corresponding to the object to be created

Suppose we want to create and delete the corresponding Dataset and AlluxioRuntime objects based on `dataset.yaml` and `runtime.yaml`.

**dataset.yaml file corresponding to the Dataset**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
```

**runtime.yaml file corresponding to the AlluxioRuntime**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 8Gi
        high: "0.95"
        low: "0.7"
```

### go client code

There are two ways to create Dataset and AlluxioRuntime with go client, one is to create Dataset and AlluixoRuntime manifest through structs with the help of fluid api, and the other is to create Dataset and AlluxioRuntime manifest directly through formatted strings. 

> Note: Only some of the necessary properties to create Dataset and Runtime objects are listed here. Dataset and AlluxioRuntime can be configured with much more than these properties. If you want to configure more properties, you can follow the example below, combined with `api/v1alpha1/dataset_types.go` and `api /v1alpha1/alluxioruntime_types.go` files, which contain code detailing the names and types of the Dataset and AlluxioRuntime properties.

#### Create Dataset and AlluxioRuntime manifest with the help of fluid api

```go
package main

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func createObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, namespace string, manifest []byte) error{
	obj := &unstructured.Unstructured{}
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	if _, _, err := decoder.Decode(manifest, &gvk, obj); err != nil {
		return err
	}
	_, err := dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func deleteObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string, name string) error{
	return dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func getObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string, name string, obj runtime.Object) error {
	data, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	dataJson, err := data.MarshalJSON()
	if err != nil {
		return err
	}

	if err = json.Unmarshal(dataJson, obj); err != nil {
		return err
	}
	return nil
}

func main() {
	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, _ := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")

	// creates the client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	datasetGVR := schema.GroupVersionResource{
		Group: "data.fluid.io",
		Version: "v1alpha1",
		Resource: "datasets",
	}
	datasetGVK := schema.GroupVersionKind{
		Group:   "data.fluid.io",
		Version: "v1alpha1",
		Kind:    "Dataset",
	}

	// Dataset manifest
	dsManifest:= &v1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "spark",
			Namespace: "default",
		},
		Spec: v1alpha1.DatasetSpec{
			Mounts: [] v1alpha1.Mount{
				{
					MountPoint:     "https://mirrors.bit.edu.cn/apache/spark/",
					Name:           "spark",
				},
			},
		},
	}

	runtimeGVR := schema.GroupVersionResource{
		Group:    "data.fluid.io",
		Version:  "v1alpha1",
		Resource: "alluxioruntimes",
	}

	runtimeGVK := schema.GroupVersionKind{
		Group:   "data.fluid.io",
		Version: "v1alpha1",
		Kind:    "AlluxioRuntime",
	}

	// AlluxioRuntime manifest
	quantity := resource.MustParse("8Gi")
	arManifest := &v1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "spark",
			Namespace: "default",
		},
		Spec: v1alpha1.AlluxioRuntimeSpec{
			Replicas: 1,
			Tieredstore: v1alpha1.Tieredstore{
				Levels: []v1alpha1.Level{
					{
						MediumType: "MEM",
						Path:       "/dev/shm",
						Quota:      &quantity,
						High:       "0.95",
						Low:        "0.7",
					},
				},
			},
		},
	}

	manifest, err := json.Marshal(dsManifest)
	if err != nil {
		panic(err)
	}

	// create the dataset
	if err = createObject(dynamicClient, datasetGVR, datasetGVK, "default", manifest); err != nil{
		panic(err)
	} else {
		fmt.Println("create the dataset successfully!")
	}

	manifest, err = json.Marshal(arManifest)
	if err != nil {
		panic(err)
	}

	// create the runtime
	if err = createObject(dynamicClient, runtimeGVR, runtimeGVK, "default", manifest); err != nil {
		panic(err)
	} else {
		fmt.Println("create the runtime successfully!")
	}

	// check whether the dataset is ready
	var ready bool = false
	var dataset v1alpha1.Dataset
	for !ready {
		// get the dataset
		if err = getObject(dynamicClient, datasetGVR, "default", "spark", &dataset); err != nil {
			panic(err)
		} else {
			status := dataset.Status.Phase
			if status == v1alpha1.BoundDatasetPhase {
				fmt.Println("the dataset is bound.")
				ready = true
			} else {
				fmt.Println("the dataset is not bound, wait 10 seconds.")
				time.Sleep(10 * time.Second)
			}
		}
	}

	// delete the runtime
	if err = deleteObject(dynamicClient, runtimeGVR, "default", "spark"); err != nil{
		panic(err)
	} else {
		fmt.Println("delete the runtime successfully!")
	}

	// delete the dataset
	if err = deleteObject(dynamicClient, datasetGVR, "default", "spark"); err != nil{
		panic(err)
	} else {
		fmt.Println("delete the dataset successfully!")
	}
}
```

#### Create Dataset and AlluxioRuntime manifest by formatted string

Note that the format of the `dsManifest` and `rtManifest` strings should be the same as the format of the yaml file, and spaces cannot be replaced by tabs!

```go
package main

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// dataset manifest
const dsManifest = `
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
`

// runtime manifest
const rtManifest = `
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 8Gi
        high: "0.95"
        low: "0.7"
`

func createObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, namespace string, manifest []byte) error{
	obj := &unstructured.Unstructured{}
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	if _, _, err := decoder.Decode(manifest, &gvk, obj); err != nil {
		return err
	}
	_, err := dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func deleteObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string, name string) error{
	return dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func getObject(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string, name string, obj runtime.Object) error {
	data, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	dataJson, err := data.MarshalJSON()
	if err != nil {
		return err
	}

	if err = json.Unmarshal(dataJson, obj); err != nil {
		return err
	}
	return nil
}

func main() {
	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, _ := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")

	// creates the client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	datasetGVR := schema.GroupVersionResource{
		Group: "data.fluid.io",
		Version: "v1alpha1",
		Resource: "datasets",
	}
	datasetGVK := schema.GroupVersionKind{
		Group:   "data.fluid.io",
		Version: "v1alpha1",
		Kind:    "Dataset",
	}
	
	runtimeGVR := schema.GroupVersionResource{
		Group:    "data.fluid.io",
		Version:  "v1alpha1",
		Resource: "alluxioruntimes",
	}

	runtimeGVK := schema.GroupVersionKind{
		Group:   "data.fluid.io",
		Version: "v1alpha1",
		Kind:    "AlluxioRuntime",
	}

	// create the dataset
	if err = createObject(dynamicClient, datasetGVR, datasetGVK, "default", []byte(dsManifest)); err != nil{
		panic(err)
	} else {
		fmt.Println("create the dataset successfully!")
	}

	// create the runtime
	if err = createObject(dynamicClient, runtimeGVR, runtimeGVK, "default", []byte(rtManifest)); err != nil {
		panic(err)
	} else {
		fmt.Println("create the runtime successfully!")
	}

	// check whether the dataset is ready
	var ready bool = false
	var dataset v1alpha1.Dataset
	for !ready {
		// get the dataset
		if err = getObject(dynamicClient, datasetGVR, "default", "spark", &dataset); err != nil {
			panic(err)
		} else {
			status := dataset.Status.Phase
			if status == v1alpha1.BoundDatasetPhase {
				fmt.Println("the dataset is bound.")
				ready = true
			} else {
				fmt.Println("the dataset is not bound, wait 10 seconds.")
				time.Sleep(10 * time.Second)
			}
		}
	}

	// delete the runtime
	if err = deleteObject(dynamicClient, runtimeGVR, "default", "spark"); err != nil{
		panic(err)
	} else {
		fmt.Println("delete the runtime successfully!")
	}

	// delete the dataset
	if err = deleteObject(dynamicClient, datasetGVR, "default", "spark"); err != nil{
		panic(err)
	} else {
		fmt.Println("delete the dataset successfully!")
	}
}
```

