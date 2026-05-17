/*
Copyright 2021 The Kruise Authors.

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

package inplaceupdate

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	utilcontainermeta "github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/containermeta"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/util"
	"gomodules.xyz/jsonpatch/v2"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
	hashutil "k8s.io/kubernetes/pkg/util/hash"

	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
)

func SetOptionsDefaults(opts *UpdateOptions) *UpdateOptions {
	if opts == nil {
		opts = &UpdateOptions{}
	}

	if opts.CalculateSpec == nil {
		opts.CalculateSpec = defaultCalculateInPlaceUpdateSpec
	}

	if opts.PatchSpecToPod == nil {
		opts.PatchSpecToPod = defaultPatchUpdateSpecToPod
	}

	if opts.CheckPodUpdateCompleted == nil {
		opts.CheckPodUpdateCompleted = DefaultCheckInPlaceUpdateCompleted
	}

	if opts.CheckContainersUpdateCompleted == nil {
		opts.CheckContainersUpdateCompleted = defaultCheckContainersInPlaceUpdateCompleted
	}

	if opts.CheckPodNeedsBeUnready == nil {
		opts.CheckPodNeedsBeUnready = defaultCheckPodNeedsBeUnready
	}

	return opts
}

// defaultPatchUpdateSpecToPod returns new pod that merges spec into old pod
func defaultPatchUpdateSpecToPod(pod *v1.Pod, spec *UpdateSpec, state *workloadv1alpha1.InPlaceUpdateState) (*v1.Pod, map[string]*v1.ResourceRequirements, error) {
	klog.V(5).InfoS("Begin to in-place update pod", "namespace", pod.Namespace, "name", pod.Name, "spec", util.DumpJSON(spec), "state", util.DumpJSON(state))

	state.NextContainerImages = make(map[string]string)
	state.NextContainerRefMetadata = make(map[string]metav1.ObjectMeta)
	state.NextContainerResources = make(map[string]v1.ResourceRequirements)

	if spec.MetaDataPatch != nil {
		cloneBytes, _ := json.Marshal(pod)
		modified, err := strategicpatch.StrategicMergePatch(cloneBytes, spec.MetaDataPatch, &v1.Pod{})
		if err != nil {
			return nil, nil, err
		}
		pod = &v1.Pod{}
		if err = json.Unmarshal(modified, pod); err != nil {
			return nil, nil, err
		}
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	// prepare containers that should update this time and next time, according to their priorities
	containersToUpdate := sets.NewString()
	var highestPriority *int
	var containersWithHighestPriority []string
	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		_, existImage := spec.ContainerImages[c.Name]
		_, existMetadata := spec.ContainerRefMetadata[c.Name]
		_, existResource := spec.ContainerResources[c.Name]
		if !existImage && !existMetadata && !existResource {
			continue
		}
		priority := getContainerPriority(c)
		if priority == nil {
			containersToUpdate.Insert(c.Name)
		} else if highestPriority == nil || *highestPriority < *priority {
			highestPriority = priority
			containersWithHighestPriority = []string{c.Name}
		} else if *highestPriority == *priority {
			containersWithHighestPriority = append(containersWithHighestPriority, c.Name)
		}
	}
	for _, cName := range containersWithHighestPriority {
		containersToUpdate.Insert(cName)
	}
	addMetadataSharedContainersToUpdate(pod, containersToUpdate, spec.ContainerRefMetadata)

	// DO NOT modify the fields in spec for it may have to retry on conflict in updatePodInPlace

	// update images and record current imageIDs for the containers to update
	containersImageChanged := sets.NewString()
	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		newImage, exists := spec.ContainerImages[c.Name]
		if !exists {
			continue
		}
		if containersToUpdate.Has(c.Name) {
			pod.Spec.Containers[i].Image = newImage
			containersImageChanged.Insert(c.Name)
		} else {
			state.NextContainerImages[c.Name] = newImage
		}
	}
	for _, c := range pod.Status.ContainerStatuses {
		if containersImageChanged.Has(c.Name) {
			if state.LastContainerStatuses == nil {
				state.LastContainerStatuses = map[string]workloadv1alpha1.InPlaceUpdateContainerStatus{}
			}
			if cs, ok := state.LastContainerStatuses[c.Name]; !ok {
				state.LastContainerStatuses[c.Name] = workloadv1alpha1.InPlaceUpdateContainerStatus{ImageID: c.ImageID}
			} else {
				// now just update imageID
				cs.ImageID = c.ImageID
			}
		}
	}

	// update annotations and labels for the containers to update
	for cName, objMeta := range spec.ContainerRefMetadata {
		if containersToUpdate.Has(cName) {
			for k, v := range objMeta.Labels {
				pod.Labels[k] = v
			}
			for k, v := range objMeta.Annotations {
				pod.Annotations[k] = v
			}
		} else {
			state.NextContainerRefMetadata[cName] = objMeta
		}
	}

	// add the containers that update this time into PreCheckBeforeNext, so that next containers can only
	// start to update when these containers have updated ready
	// TODO: currently we only support ContainersRequiredReady, not sure if we have to add ContainersPreferredReady in future
	if len(state.NextContainerImages) > 0 || len(state.NextContainerRefMetadata) > 0 || len(state.NextContainerResources) > 0 {
		state.PreCheckBeforeNext = &workloadv1alpha1.InPlaceUpdatePreCheckBeforeNext{ContainersRequiredReady: containersToUpdate.List()}
	} else {
		state.PreCheckBeforeNext = nil
	}

	state.ContainerBatchesRecord = append(state.ContainerBatchesRecord, workloadv1alpha1.InPlaceUpdateContainerBatch{
		Timestamp:  metav1.NewTime(Clock.Now()),
		Containers: containersToUpdate.List(),
	})

	klog.V(5).InfoS("Decide to in-place update pod", "namespace", pod.Namespace, "name", pod.Name, "state", util.DumpJSON(state))

	inPlaceUpdateStateJSON, _ := json.Marshal(state)
	pod.Annotations[workloadv1alpha1.InPlaceUpdateStateKey] = string(inPlaceUpdateStateJSON)
	return pod, nil, nil
}

func addMetadataSharedContainersToUpdate(pod *v1.Pod, containersToUpdate sets.String, containerRefMetadata map[string]metav1.ObjectMeta) {
	labelsToUpdate := sets.NewString()
	annotationsToUpdate := sets.NewString()
	newToUpdate := containersToUpdate
	// We need a for-loop to merge the indirect shared containers
	for newToUpdate.Len() > 0 {
		for _, cName := range newToUpdate.UnsortedList() {
			if objMeta, exists := containerRefMetadata[cName]; exists {
				for key := range objMeta.Labels {
					labelsToUpdate.Insert(key)
				}
				for key := range objMeta.Annotations {
					annotationsToUpdate.Insert(key)
				}
			}
		}
		newToUpdate = sets.NewString()

		for cName, objMeta := range containerRefMetadata {
			if containersToUpdate.Has(cName) {
				continue
			}
			for _, key := range labelsToUpdate.UnsortedList() {
				if _, exists := objMeta.Labels[key]; exists {
					klog.InfoS("Has to in-place update container with lower priority in Pod, for the label it shared has changed",
						"containerName", cName, "namespace", pod.Namespace, "name", pod.Name, "label", key)
					containersToUpdate.Insert(cName)
					newToUpdate.Insert(cName)
					break
				}
			}
			for _, key := range annotationsToUpdate.UnsortedList() {
				if _, exists := objMeta.Annotations[key]; exists {
					klog.InfoS("Has to in-place update container with lower priority in Pod, for the annotation it shared has changed",
						"containerName", cName, "namespace", pod.Namespace, "podName", pod.Name, "annotation", key)
					containersToUpdate.Insert(cName)
					newToUpdate.Insert(cName)
					break
				}
			}
		}
	}
}

// defaultCalculateInPlaceUpdateSpec calculates diff between old and update revisions.
// If the diff just contains replace operation of spec.containers[x].image, it will returns an UpdateSpec.
// Otherwise, it returns nil which means can not use in-place update.
func defaultCalculateInPlaceUpdateSpec(oldRevision, newRevision *apps.ControllerRevision, opts *UpdateOptions) *UpdateSpec {
	if oldRevision == nil || newRevision == nil {
		return nil
	}
	opts = SetOptionsDefaults(opts)

	patches, err := jsonpatch.CreatePatch(oldRevision.Data.Raw, newRevision.Data.Raw)
	if err != nil {
		return nil
	}

	oldTemp, err := GetTemplateFromRevision(oldRevision)
	if err != nil {
		return nil
	}
	newTemp, err := GetTemplateFromRevision(newRevision)
	if err != nil {
		return nil
	}

	updateSpec := &UpdateSpec{
		Revision:             newRevision.Name,
		ContainerImages:      make(map[string]string),
		ContainerResources:   make(map[string]v1.ResourceRequirements),
		ContainerRefMetadata: make(map[string]metav1.ObjectMeta),
		GraceSeconds:         opts.GracePeriodSeconds,
	}
	if opts.GetRevision != nil {
		updateSpec.Revision = opts.GetRevision(newRevision)
	}

	// all patches for podSpec can just update images in pod spec
	var metadataPatches []jsonpatch.Operation
	for _, op := range patches {
		op.Path = strings.Replace(op.Path, "/spec/template", "", 1)

		if !strings.HasPrefix(op.Path, "/spec/") {
			if strings.HasPrefix(op.Path, "/metadata/") {
				metadataPatches = append(metadataPatches, op)
				continue
			}
			return nil
		}

		if op.Operation != "replace" {
			return nil
		}
		if containerImagePatchRexp.MatchString(op.Path) {
			// for example: /spec/containers/0/image
			words := strings.Split(op.Path, "/")
			idx, _ := strconv.Atoi(words[3])
			if len(oldTemp.Spec.Containers) <= idx {
				return nil
			}
			updateSpec.ContainerImages[oldTemp.Spec.Containers[idx].Name] = op.Value.(string)
			continue
		}

		return nil
	}

	if len(metadataPatches) > 0 {
		oldBytes, _ := json.Marshal(v1.Pod{ObjectMeta: oldTemp.ObjectMeta})
		newBytes, _ := json.Marshal(v1.Pod{ObjectMeta: newTemp.ObjectMeta})
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldBytes, newBytes, &v1.Pod{})
		if err != nil {
			return nil
		}
		updateSpec.MetaDataPatch = patchBytes
	}

	return updateSpec
}

// DefaultCheckInPlaceUpdateCompleted checks whether imageID in pod status has been changed since in-place update.
// If the imageID in containerStatuses has not been changed, we assume that kubelet has not updated
// containers in Pod.
func DefaultCheckInPlaceUpdateCompleted(pod *v1.Pod) error {
	if _, isInGraceState := workloadv1alpha1.GetInPlaceUpdateGrace(pod); isInGraceState {
		return fmt.Errorf("still in grace period of in-place update")
	}

	inPlaceUpdateState := workloadv1alpha1.InPlaceUpdateState{}
	if stateStr, ok := workloadv1alpha1.GetInPlaceUpdateState(pod); !ok {
		return nil
	} else if err := json.Unmarshal([]byte(stateStr), &inPlaceUpdateState); err != nil {
		return err
	}
	if len(inPlaceUpdateState.NextContainerImages) > 0 || len(inPlaceUpdateState.NextContainerRefMetadata) > 0 || len(inPlaceUpdateState.NextContainerResources) > 0 {
		return fmt.Errorf("existing containers to in-place update in next batches")
	}
	return defaultCheckContainersInPlaceUpdateCompleted(pod, &inPlaceUpdateState)
}

func defaultCheckContainersInPlaceUpdateCompleted(pod *v1.Pod, inPlaceUpdateState *workloadv1alpha1.InPlaceUpdateState) error {
	runtimeContainerMetaSet, err := workloadv1alpha1.GetRuntimeContainerMetaSet(pod)
	if err != nil {
		return err
	}

	if inPlaceUpdateState.UpdateEnvFromMetadata {
		if runtimeContainerMetaSet == nil {
			return fmt.Errorf("waiting for all containers hash consistent, but runtime-container-meta not found")
		}
		if !checkAllContainersHashConsistent(pod, runtimeContainerMetaSet, extractedEnvFromMetadataHash) {
			return fmt.Errorf("waiting for all containers hash consistent")
		}
	}

	// only UpdateResources, we check resources in status updated

	if runtimeContainerMetaSet != nil {
		metaHashType := plainHash
		if checkAllContainersHashConsistent(pod, runtimeContainerMetaSet, metaHashType) {
			klog.V(5).InfoS("Check Pod in-place update completed for all container hash consistent", "namespace", pod.Namespace, "name", pod.Name)
			return nil
		}
		// If it needs not to update envs from metadata, we don't have to return error here,
		// in case kruise-daemon has broken for some reason and runtime-container-meta is still in an old version.
	}

	containerImages := make(map[string]string, len(pod.Spec.Containers))
	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		containerImages[c.Name] = c.Image
		if len(strings.Split(c.Image, ":")) <= 1 {
			containerImages[c.Name] = fmt.Sprintf("%s:latest", c.Image)
		}
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if oldStatus, ok := inPlaceUpdateState.LastContainerStatuses[cs.Name]; ok {
			// TODO: we assume that users should not update workload template with new image which actually has the same imageID as the old image
			if oldStatus.ImageID == cs.ImageID {
				if containerImages[cs.Name] != cs.Image {
					return fmt.Errorf("container %s imageID not changed", cs.Name)
				}
			}
			delete(inPlaceUpdateState.LastContainerStatuses, cs.Name)
		}
	}

	if len(inPlaceUpdateState.LastContainerStatuses) > 0 {
		return fmt.Errorf("not found statuses of containers %v", inPlaceUpdateState.LastContainerStatuses)
	}

	return nil
}

type hashType string

const (
	plainHash                    hashType = "PlainHash"
	extractedEnvFromMetadataHash hashType = "ExtractedEnvFromMetadataHash"
)

// The requirements for hash consistent:
// 1. all containers in spec.containers should also be in status.containerStatuses and runtime-container-meta
// 2. all containers in status.containerStatuses and runtime-container-meta should have the same containerID
// 3. all containers in spec.containers and runtime-container-meta should have the same hashes
func checkAllContainersHashConsistent(pod *v1.Pod, runtimeContainerMetaSet *workloadv1alpha1.RuntimeContainerMetaSet, hashType hashType) bool {
	for i := range pod.Spec.Containers {
		containerSpec := &pod.Spec.Containers[i]

		var containerStatus *v1.ContainerStatus
		for j := range pod.Status.ContainerStatuses {
			if pod.Status.ContainerStatuses[j].Name == containerSpec.Name {
				containerStatus = &pod.Status.ContainerStatuses[j]
				break
			}
		}
		if containerStatus == nil {
			klog.InfoS("Find no container in status for Pod", "containerName", containerSpec.Name, "namespace", pod.Namespace, "podName", pod.Name)
			return false
		}

		var containerMeta *workloadv1alpha1.RuntimeContainerMeta
		for i := range runtimeContainerMetaSet.Containers {
			if runtimeContainerMetaSet.Containers[i].Name == containerSpec.Name {
				containerMeta = &runtimeContainerMetaSet.Containers[i]
				continue
			}
		}
		if containerMeta == nil {
			klog.InfoS("Find no container in runtime-container-meta for Pod", "containerName", containerSpec.Name, "namespace", pod.Namespace, "podName", pod.Name)
			return false
		}

		if containerMeta.ContainerID != containerStatus.ContainerID {
			klog.InfoS("Find container in runtime-container-meta for Pod has different containerID with status",
				"containerName", containerSpec.Name, "namespace", pod.Namespace, "podName", pod.Name,
				"metaID", containerMeta.ContainerID, "statusID", containerStatus.ContainerID)
			return false
		}

		switch hashType {
		case plainHash:
			isConsistentInNewVersion := hashContainer(containerSpec) == containerMeta.Hashes.PlainHash
			isConsistentInOldVersion := hashContainer(containerSpec) == containerMeta.Hashes.PlainHash
			if !isConsistentInNewVersion && !isConsistentInOldVersion {
				klog.InfoS("Find container in runtime-container-meta for Pod has different plain hash with spec",
					"containerName", containerSpec.Name, "namespace", pod.Namespace, "podName", pod.Name,
					"metaHash", containerMeta.Hashes.PlainHash, "expectedHashInNewVersion", hashContainer(containerSpec), "expectedHashInOldVersion", hashContainer(containerSpec))
				return false
			}
		case extractedEnvFromMetadataHash:
			hasher := utilcontainermeta.NewEnvFromMetadataHasher()
			if expectedHash := hasher.GetExpectHash(containerSpec, pod); containerMeta.Hashes.ExtractedEnvFromMetadataHash != expectedHash {
				klog.InfoS("Find container in runtime-container-meta for Pod has different extractedEnvFromMetadataHash with spec",
					"containerName", containerSpec.Name, "namespace", pod.Namespace, "podName", pod.Name,
					"metaHash", containerMeta.Hashes.ExtractedEnvFromMetadataHash, "expectedHash", expectedHash)
				return false
			}
		}
	}

	return true
}

// hashContainer copy from kubelet v1.31-
// in 1.31+, kubeletcontainer.HashContainer will only pick some fields to hash
// in order to be compatible with 1.31 and earlier, here is the implementation of kubeletcontainer.HashContainer(1.31-) copied.
func hashContainer(container *v1.Container) uint64 {
	hash := fnv.New32a()
	// Omit nil or empty field when calculating hash value
	// Please see https://github.com/kubernetes/kubernetes/issues/53644
	containerJSON, _ := json.Marshal(container)
	hashutil.DeepHashObject(hash, containerJSON)
	return uint64(hash.Sum32())
}

const (
	cpuMask = 1
	memMask = 2
)

func defaultCheckPodNeedsBeUnready(pod *v1.Pod, spec *UpdateSpec) bool {
	return containsReadinessGate(pod)
}

// getContainerPriority returns the container launch priority if set via ContainerLaunchBarrierEnvName env.
func getContainerPriority(c *v1.Container) *int {
	const priorityStartIndex = 2
	for _, e := range c.Env {
		if e.Name == workloadv1alpha1.ContainerLaunchBarrierEnvName {
			p, _ := strconv.Atoi(e.ValueFrom.ConfigMapKeyRef.Key[priorityStartIndex:])
			return &p
		}
	}
	return nil
}
