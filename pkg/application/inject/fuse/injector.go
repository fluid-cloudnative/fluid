/*
Copyright 2022 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fuse

import (
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/mutator"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/defaultapp"
	podapp "github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/unstructured"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredtype "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Injector struct {
	client client.Client
	log    logr.Logger
}

func NewInjector(client client.Client) *Injector {
	return &Injector{
		client: client,
		log:    ctrl.Log.WithName("fuse-injector"),
	}
}

// InjectPod injects pod with runtimeInfo which key is pvcName, value is runtimeInfo
func (s *Injector) InjectPod(in *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (out *corev1.Pod, err error) {
	match := false
	outObj, err := s.Inject(in, runtimeInfos)
	if err != nil {
		return out, err
	}
	out, match = outObj.(*corev1.Pod)
	if !match {
		return out, fmt.Errorf("failed to match the object %v to %v", outObj, out)
	}

	return
}

// inject takes the following steps to inject fuse container:
// 1. Determine the type of the input runtime.Object and wrap it as a FluidApplication which contains one or more PodSpecs.
// 2. For each PodSpec in the FluidApplication, and for each runtimeInfo involved, inject the PodSpec according to the runtimeInfo(i.e. func `injectObject()`)
// 3. Add injection done label to the PodSpec, indicating mutation is done for the PodSpec.
// 4. When all the PodSpecs are mutated, return the modified runtime.Object
func (s *Injector) inject(in runtime.Object, runtimeInfos map[string]base.RuntimeInfoInterface) (out runtime.Object, err error) {
	out = in.DeepCopyObject()

	var (
		application common.FluidApplication
		objectMeta  metav1.ObjectMeta
		typeMeta    metav1.TypeMeta
	)

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

			r, err := s.inject(obj, runtimeInfos)
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
		application = podapp.NewApplication(pod)
	case *unstructuredtype.Unstructured:
		obj := v
		typeMeta = metav1.TypeMeta{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
		}
		objectMeta = metav1.ObjectMeta{
			Name:        obj.GetName(),
			Namespace:   obj.GetNamespace(),
			Annotations: obj.GetAnnotations(),
			Labels:      obj.GetLabels(),
		}
		application = unstructured.NewApplication(obj)
	case runtime.Object:
		obj := v
		outValue := reflect.ValueOf(obj).Elem()
		typeMeta = outValue.FieldByName("TypeMeta").Interface().(metav1.TypeMeta)
		objectMeta = outValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
		application = defaultapp.NewApplication(obj)
	default:
		s.log.Info("No supported K8s Type", "v", v)
		return out, fmt.Errorf("no supported K8s Type %v", v)
	}

	s.log.V(1).Info("Inject Fluid application", "objectMeta.Name", objectMeta.Name, "objectMeta.GenerateName", objectMeta.GenerateName, "objectMeta.Namespace", objectMeta.Namespace, "kind", typeMeta.Kind)
	pods, err := application.GetPodSpecs()
	if err != nil {
		return
	}

	for _, pod := range pods {
		podObjMeta, err := pod.GetMetaObject()
		if err != nil {
			s.log.Error(err, "failed to getMetaObject of pod", "fluid application name", objectMeta.Name)
			return out, err
		}

		// Pod may have either Name or GenerateName set, take it as podName for log messages
		podName := podObjMeta.Name
		if len(podName) == 0 {
			podName = podObjMeta.GenerateName
		}

		shouldInject, err := s.shouldInject(pod)
		if err != nil {
			s.log.Error(err, "failed to check if should inject pod", "pod name", podName)
			return out, err
		}

		if !shouldInject {
			continue
		}

		if err = s.injectCheckMountReadyScript(pod, runtimeInfos); err != nil {
			s.log.Error(err, "failed to injectCheckMountReadyScript()", "pod name", podName)
			return out, err
		}

		idx := 0
		for pvcName, runtimeInfo := range runtimeInfos {
			s.log.Info("Start mutating pvc in pod spec", "pod name", podName, "pvc name", pvcName)
			// Append no suffix to fuse container name unless there are multiple ones.
			containerNameSuffix := fmt.Sprintf("-%d", idx)

			if err = s.injectObject(pod, pvcName, runtimeInfo, containerNameSuffix); err != nil {
				s.log.Error(err, "failed to injectObject()", "pod name", podName, "pvc name", pvcName)
				return out, err
			}

			idx++
		}

		if err = s.labelInjectionDone(pod); err != nil {
			s.log.Error(err, "failed to labelInjectionDone()", "pod name", podName)
			return out, err
		}
	}

	// kubeclient.IsVolumeMountForPVC(pvcName, )

	err = application.SetPodSpecs(pods)
	if err != nil {
		return out, err
	}

	out = application.GetObject()
	return out, nil
}

// injectObject injects fuse container into a PodSpec given the pvcName and the runtimeInfo. It takes the following steps:
// 1. Check if it needs injection by checking PodSpec's original information.
// 2. Generate the fuse container template to inject.
// 3. Handle mutations on the PodSpec's volumes
// 4. Handle mutations on the PodSpec's volumeMounts
// 5. Add the fuse container to the first of the PodSpec's container list
func (s *Injector) injectObject(pod common.FluidObject, pvcName string, runtimeInfo base.RuntimeInfoInterface, containerNameSuffix string) (err error) {
	// Cannot use objMeta.namespace as the expected namespace because it may be empty and not trustworthy before Kubernetes 1.24.
	// For more details, see https://github.com/kubernetes/website/issues/30574#issuecomment-974896246
	specsToMutate, err := mutator.CollectFluidObjectSpecs(pod)
	if err != nil {
		return err
	}

	mountedPvc := kubeclient.PVCNames(specsToMutate.VolumeMounts, specsToMutate.Volumes)
	found := utils.ContainsString(mountedPvc, pvcName)
	if !found {
		s.log.Info("Not able to find the fluid pvc in pod spec, skip",
			"pvc", pvcName,
			"mounted pvcs", mountedPvc)
		return nil
	}

	option := common.FuseSidecarInjectOption{
		EnableCacheDir:             utils.InjectCacheDirEnabled(specsToMutate.MetaObj.Labels),
		EnableUnprivilegedSidecar:  utils.FuseSidecarUnprivileged(specsToMutate.MetaObj.Labels),
		SkipSidecarPostStartInject: utils.SkipSidecarPostStartInject(specsToMutate.MetaObj.Labels),
	}

	// template, exist := cache.GetFuseTemplateByKey(pvcKey, option)
	// if !exist {
	// 	template, err = runtimeInfo.GetTemplateToInjectForFuse(pvcName, pvcNamespace, option)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	cache.AddFuseTemplateByKey(pvcKey, option, template)
	// }

	// TODO: support cache for faster path to get fuse container template.
	template, err := runtimeInfo.GetFuseContainerTemplate()
	if err != nil {
		return err
	}

	mutatorBuildOpts := mutator.MutatorBuildOpts{
		PvcName:     pvcName,
		Template:    template,
		Options:     option,
		RuntimeInfo: runtimeInfo,
		NameSuffix:  containerNameSuffix,
		Client:      s.client,
		Log:         s.log,
		Specs:       specsToMutate,
	}

	platform := s.getServerlessPlatformFromMeta(specsToMutate.MetaObj)
	if len(platform) == 0 {
		return fmt.Errorf("can't find any supported platform-specific mutator in pod's metadata")
	}

	mtt, err := mutator.BuildMutator(mutatorBuildOpts, platform)
	if err != nil {
		return err
	}

	if err := mtt.PrepareMutation(); err != nil {
		return err
	}

	mutatedSpecs, err := mtt.Mutate()
	if err != nil {
		return err
	}

	if err := mutator.ApplyFluidObjectSpecs(pod, mutatedSpecs); err != nil {
		return err
	}

	return nil
}

// Inject delegates inject() to do all the mutations
func (s *Injector) Inject(in runtime.Object, runtimeInfos map[string]base.RuntimeInfoInterface) (out runtime.Object, err error) {
	return s.inject(in, runtimeInfos)
}

func (s *Injector) InjectUnstructured(in *unstructuredtype.Unstructured, runtimeInfos map[string]base.RuntimeInfoInterface) (out *unstructuredtype.Unstructured, err error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *Injector) shouldInject(pod common.FluidObject) (should bool, err error) {
	metaObj, err := pod.GetMetaObject()
	if err != nil {
		return should, err
	}

	// Skip if pod does not enable serverless injection (i.e. lack of specific label)
	if !utils.ServerlessEnabled(metaObj.Labels) || utils.InjectSidecarDone(metaObj.Labels) {
		s.log.V(1).Info("Serverless injection not enabled in pod labels, skip")
		return should, nil
	}

	// Skip if found existing container with conflicting name.
	allContainerNames, err := collectAllContainerNames(pod)
	if err != nil {
		return should, err
	}
	for _, cName := range allContainerNames {
		if cName == common.FuseContainerName || cName == common.InitFuseContainerName {
			s.log.Info("Found existing conflict container name before injection, skip", "containerName", cName)
			return should, nil
		}
	}

	should = true
	return should, nil
}

// labelInjectionDone adds a injecting done label to a PodSpec, indicating all the mutations have been finished
func (s *Injector) labelInjectionDone(pod common.FluidObject) error {
	metaObj, err := pod.GetMetaObject()
	if err != nil {
		return err
	}

	if metaObj.Labels == nil {
		metaObj.Labels = map[string]string{}
	}

	metaObj.Labels[common.InjectSidecarDone] = common.True
	err = pod.SetMetaObject(metaObj)
	if err != nil {
		return err
	}

	return nil
}

func (s *Injector) getServerlessPlatformFromMeta(metaObj metav1.ObjectMeta) string {
	return utils.GetServerlessPlatfrom(metaObj.Labels)
}
