package serverless

import (
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/object"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/unstructured"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredtype "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Injector struct {
	client           client.Client
	defaultNamespace string
}

func NewInjector(client client.Client, namespace string) *Injector {
	return &Injector{
		client:           client,
		defaultNamespace: namespace,
	}
}

func (s *Injector) createTemplate(datasetName, namespace string) (template *serverlessInjectionTemplate, err error) {
	return &serverlessInjectionTemplate{}, nil
}

func (s *Injector) InjectObject(in runtime.Object) (out runtime.Object, err error) {
	out = in.DeepCopyObject()

	var application common.Application
	var objectMeta metav1.ObjectMeta
	var typeMeta metav1.TypeMeta
	var namespace string

	// Handle Lists
	if list, ok := out.(*corev1.List); ok {
		result := list

		for i, item := range list.Items {
			obj, err := utils.FromRawToObject(item.Raw)
			if runtime.IsNotRegisteredError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			r, err := s.InjectObject(obj)
			if err != nil {
				return nil, err
			}

			re := runtime.RawExtension{}
			re.Object = r
			result.Items[i] = re
		}
		return result, nil
	}

	switch v := out.(type) {
	case *corev1.Pod:
		pod := v
		typeMeta = pod.TypeMeta
		objectMeta = pod.ObjectMeta
		application = object.NewRuntimeApplication(pod)
	case *unstructuredtype.Unstructured:
		obj := v
		application = unstructured.NewUnstructuredApplication(obj)
		typeMeta = metav1.TypeMeta{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
		}
		objectMeta = metav1.ObjectMeta{
			Name:        obj.GetName(),
			Namespace:   obj.GetNamespace(),
			Annotations: obj.GetAnnotations(),
		}
	case runtime.Object:
		obj := v
		outValue := reflect.ValueOf(obj).Elem()
		application = object.NewRuntimeApplication(obj)
		typeMeta = outValue.FieldByName("TypeMeta").Interface().(metav1.TypeMeta)
		objectMeta = outValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)

	default:
		log.Info("No supported K8s Type", "v", v)
		return out, fmt.Errorf("No supported K8s Type", "v", v)
	}

	isServerless := ServerlessEnabled(objectMeta.Annotations)
	if !isServerless {
		return out, nil
	}

	if len(objectMeta.Namespace) == 0 {
		namespace = s.defaultNamespace
	}

	name := types.NamespacedName{
		Namespace: namespace,
		Name:      objectMeta.Name,
	}
	kind := typeMeta.Kind

	pods, err := application.GetPodSpecs()
	if err != nil {
		return
	}

	for _, pod := range pods {
		volumes, err := pod.GetVolumes()
		if err != nil {
			return out, err
		}
		for _, volume := range volumes {

			dataset, err := kubeclient.GetDatasetFromPVC(s.client, volume.PersistentVolumeClaim.ClaimName, namespace)
			if err != nil {
				return out, err
			}

			if dataset != nil {
				template, err := s.createTemplate(dataset.Name, dataset.Namespace)
				if err != nil {
					return out, err
				}

				containers, err := pod.GetContainers()
				if err != nil {
					return out, err
				}

				// Skip injection for injected container
				for _, c := range containers {
					if c.Name == fuseContainerName {
						warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
							name, fuseContainerName)
						if len(kind) != 0 {
							warningStr = fmt.Sprintf("===> Skipping injection because Kind %s: %v has injected %q sidecar already\n",
								kind, name, fuseContainerName)
						}
						log.Info(warningStr)
						return out, nil
					}
				}

				// 1.Modify the volumes
				volumes, err := pod.GetVolumes()
				if err != nil {
					return out, err
				}

				for i, v := range volumes {
					for _, toUpdate := range template.VolumesToUpdate {
						if v.Name == toUpdate.Name {
							log.V(1).Info("Update volume", "original", v, "updated", toUpdate)
							volumes[i] = toUpdate
							// break
						}
					}
				}

				// 2.Add the volumes
				if len(template.VolumesToAdd) > 0 {
					log.V(1).Info("Before append volume", "original", volumes)
					volumes = append(volumes, template.VolumesToAdd...)
					log.V(1).Info("After append volume", "original", volumes)
				}

				// 3.Add sidecar as the first container
				containers = append([]corev1.Container{template.FuseContainer}, containers...)

				err = pod.SetContainers(containers)
				if err != nil {
					return out, err
				}

				err = pod.SetVolumes(volumes)
				if err != nil {
					return out, err
				}

			}
		}

	}

	err = application.SetPodSpecs(pods)
	if err != nil {
		return out, err
	}

	out = application.GetObject()
	return out, err
}
