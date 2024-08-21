package mounteddatasetinjector

import (
	"fmt"
	"slices"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name = "MountedDatasetInjector"

var (
	log = ctrl.Log.WithName(Name)
)

type MountedDatasetInjector struct {
	client client.Client
	name   string
}

var _ api.MutatingHandler = &MountedDatasetInjector{}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &MountedDatasetInjector{
		client: c,
		name:   Name,
	}, nil
}

// TODO: Support cases where fuse sidecars are injected in multi-round. Currently, only dataset names in the first round will be recorded.
func (injector *MountedDatasetInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	if len(runtimeInfos) == 0 {
		return false, nil
	}

	podName := pod.Name
	if len(pod.Name) == 0 {
		podName = pod.GenerateName
	}

	mountedDatasets := make([]string, 0, len(runtimeInfos))
	for _, runtimeInfo := range runtimeInfos {
		mountedDatasets = append(mountedDatasets, runtimeInfo.GetName())
	}
	slices.Sort(mountedDatasets)
	log.Info("Injecting mounted dataset annotation to pod",
		"annotation", fmt.Sprintf("%s=%s", common.LabelAnnotationMountedDatasets, strings.Join(mountedDatasets, ",")),
		"pod", fmt.Sprintf("%s/%s", pod.Namespace, podName))

	if len(pod.Annotations) == 0 {
		pod.Annotations = map[string]string{}
	}

	if val, exists := pod.Annotations[common.LabelAnnotationMountedDatasets]; !exists || val != strings.Join(mountedDatasets, ",") {
		pod.Annotations[common.LabelAnnotationMountedDatasets] = strings.Join(mountedDatasets, ",")
	}

	return false, nil
}

func (injector *MountedDatasetInjector) GetName() string {
	return injector.name
}
