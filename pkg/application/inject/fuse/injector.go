/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
// 2. For each PodSpec in the FluidApplication, and for each runtimeInfo involved, inject the PodSpec according to the runtimeInfo and the serverless platform (i.e. func `MutateWithRuntimeInfo()`)
// 3. Pod-level mutation according to the serverless platform (i.e. func `PostMutate()`)
// 4. Add injection done label to the PodSpec, indicating mutation is done for the PodSpec.
// 5. When all the PodSpecs are mutated, return the modified runtime.Object
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
		podSpecs, err := mutator.CollectFluidObjectSpecs(pod)
		if err != nil {
			s.log.Error(err, "failed to collect fluid object specs from a pod", "fluid application name", objectMeta.Name)
		}

		// Pod may have either Name or GenerateName set, take it as podName for log messages
		podName := podSpecs.MetaObj.Name
		if len(podName) == 0 {
			podName = podSpecs.MetaObj.GenerateName
		}

		shouldInject, err := s.shouldInject(pod)
		if err != nil {
			s.log.Error(err, "failed to check if should inject pod", "pod name", podName)
			return out, err
		}

		if !shouldInject {
			continue
		}

		platform := s.getServerlessPlatformFromMeta(podSpecs.MetaObj)
		if len(platform) == 0 {
			return out, fmt.Errorf("can't find any supported platform-specific mutator in pod's metadata")
		}

		mutatorBuildOpts := mutator.MutatorBuildOpts{
			Client: s.client,
			Log:    s.log,
			Specs:  podSpecs,
			Options: common.FuseSidecarInjectOption{
				EnableCacheDir:             utils.InjectCacheDirEnabled(podSpecs.MetaObj.Labels),
				EnableUnprivilegedSidecar:  utils.FuseSidecarUnprivileged(podSpecs.MetaObj.Labels),
				SkipSidecarPostStartInject: utils.SkipSidecarPostStartInject(podSpecs.MetaObj.Labels),
			},
		}

		mtt, err := mutator.BuildMutator(mutatorBuildOpts, platform)
		if err != nil {
			return out, err
		}

		if err = s.injectCheckMountReadyScript(podSpecs, runtimeInfos); err != nil {
			s.log.Error(err, "failed to injectCheckMountReadyScript()", "pod name", podName)
			return out, err
		}

		idx := 0
		for pvcName, runtimeInfo := range runtimeInfos {
			s.log.Info("Start mutating pvc in pod spec", "pod name", podName, "pvc name", pvcName)
			// Append no suffix to fuse container name unless there are multiple ones.
			containerNameSuffix := fmt.Sprintf("-%d", idx)

			if err = mtt.MutateWithRuntimeInfo(pvcName, runtimeInfo, containerNameSuffix); err != nil {
				s.log.Error(err, "failed to mutate pod for the pvc", "pod name", podName, "pvc name", pvcName)
				return out, err
			}

			idx++
		}

		if err = mtt.PostMutate(); err != nil {
			s.log.Error(err, "failed to execute PostMutate() for the pod", "pod name", podName)
			return out, err
		}

		if err = mutator.ApplyFluidObjectSpecs(pod, mtt.GetMutatedPodSpecs()); err != nil {
			s.log.Error(err, "error when applying mutated specs to pod", "pod name", podName)
			return out, err
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
