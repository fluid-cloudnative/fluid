# 如何使用Go客户端创建、删除fluid资源

下面以创建和删除Dataset和AlluxioRuntime为例，展示如何通过使用Go客户端创建、删除fluid资源。

### 要创建的对象对应的yaml文件

假定我们要依据`dataset.yaml`和`runtime.yaml`创建和删除相应的Dataset和AlluxioRuntime对象。

**Dataset对应的dataset.yaml文件**

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

**AlluxioRuntime对应的runtime.yaml文件**

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

### go客户端代码

采用go客户端创建Dataset和AlluxioRuntime。这里有两种方式，一种是借助fluid的api通过结构体来创建Dataset和AlluixoRuntime的manifest，另一种方式是直接通过格式化的字符串来创建Dataset和AlluxioRuntime的manifest。

>注意：这里只列举了创建Dataset和Runtime对象的部分必要属性，Dataset和AlluxioRuntime可配置的属性远不只这些，如果想要配置其它更多属性，可仿照以下例子，结合`api/v1alpha1/dataset_types.go`以及`api/v1alpha1/alluxioruntime_types.go`文件进行配置，这两个文件中的代码详细写明了Dataset和AlluxioRuntime各个属性的名称及类型。

#### 借助fluid的api来创建Dataset和AlluxioRuntime的manifest

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

#### 通过格式化的字符串来创建Dataset和AlluxioRuntime的manifest

注意`dsManifest`和`rtManifest`字符串的格式应该是和yaml文件的格式一致的,空格不能用tab代替!

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

