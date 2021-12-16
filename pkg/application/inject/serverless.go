package inject

import (
	"fmt"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	injectScheme      = runtime.NewScheme()
	fuseContainerName = "fluid-fuse"
)

type (
	Template  corev1.Pod
	Templates map[string]string
)

// InjectObject injects sidecar into
func InjectObject(in runtime.Object, sidecarTemplate Templates) (out interface{}, err error) {
	out = in.DeepCopyObject()

	var containers []corev1.Container
	var volumes []corev1.Volume
	var metadata *metav1.ObjectMeta
	var typeMeta metav1.TypeMeta

	// Handle Lists
	if list, ok := out.(*corev1.List); ok {
		result := list

		for i, item := range list.Items {
			obj, err := fromRawToObject(item.Raw)
			if runtime.IsNotRegisteredError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			r, err := InjectObject(obj, sidecarTemplate)
			if err != nil {
				return nil, err
			}

			re := runtime.RawExtension{}
			re.Object = r.(runtime.Object)
			result.Items[i] = re
		}
		return result, nil
	}

	switch v := out.(type) {
	case *corev1.Pod:
		pod := v
		containers = pod.Spec.Containers
		volumes = pod.Spec.Volumes
		typeMeta = pod.TypeMeta
		metadata = &pod.ObjectMeta
	default:
		log.Info("No supported K8s type", "v", v)
		return out, fmt.Errorf("No support for k8s type", "v", v)
	}

	name := types.NamespacedName{
		Namespace: metadata.Namespace,
		Name:      metadata.Name,
	}
	// Skip injection for injected container
	if len(containers) > 0 {
		for _, c := range containers {
			if c.Name == fuseContainerName {
				warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
					name, fuseContainerName)
				log.Info(warningStr)
				return out, nil
			}
		}
	}

	return out, err
}

// fromRawToObject is used to convert from raw to the runtime object
func fromRawToObject(raw []byte) (runtime.Object, error) {
	var typeMeta metav1.TypeMeta
	if err := yaml.Unmarshal(raw, &typeMeta); err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(typeMeta.APIVersion, typeMeta.Kind)
	obj, err := injectScheme.New(gvk)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(raw, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
