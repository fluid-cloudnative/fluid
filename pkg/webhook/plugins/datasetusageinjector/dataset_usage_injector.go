package datasetusageinjector

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

const Name = "DatasetUsageInjector"

var (
	log = ctrl.Log.WithName(Name)
)

type DatasetUsageInjector struct {
	client client.Client
	name   string
}

var _ api.MutatingHandler = &DatasetUsageInjector{}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &DatasetUsageInjector{
		client: c,
		name:   Name,
	}, nil
}

func (injector *DatasetUsageInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	if len(runtimeInfos) == 0 {
		return false, nil
	}

	podName := pod.Name
	if len(pod.Name) == 0 {
		podName = pod.GenerateName
	}

	annotationKey := common.LabelAnnotationDatasetsInUse

	datasetsInUseMap := make(map[string]struct{})

	// 1. Read existing datasets from annotation
	if existingVal, exists := pod.Annotations[annotationKey]; exists && len(existingVal) > 0 {
		existingDatasets := strings.Split(existingVal, ",")
		for _, ds := range existingDatasets {
			trimmed := strings.TrimSpace(ds)
			if trimmed != "" {
				datasetsInUseMap[trimmed] = struct{}{}
			}
		}
	}

	// 2. Add new datasets from current round
	for _, runtimeInfo := range runtimeInfos {
		datasetsInUseMap[runtimeInfo.GetName()] = struct{}{}
	}

	// 3. Convert map to sorted slice
	datasetsInUse := make([]string, 0, len(datasetsInUseMap))
	for ds := range datasetsInUseMap {
		datasetsInUse = append(datasetsInUse, ds)
	}
	slices.Sort(datasetsInUse)

	annotationValue := strings.Join(datasetsInUse, ",")

	log.Info("Injecting dataset usage annotation to pod",
		"annotation", fmt.Sprintf("%s=%s", annotationKey, annotationValue),
		"pod", fmt.Sprintf("%s/%s", pod.Namespace, podName))

	if len(pod.Annotations) == 0 {
		pod.Annotations = map[string]string{}
	}

	if val, exists := pod.Annotations[annotationKey]; !exists || val != annotationValue {
		pod.Annotations[annotationKey] = annotationValue
	}

	return false, nil
}

func (injector *DatasetUsageInjector) GetName() string {
	return injector.name
}
