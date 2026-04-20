package engine

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// getDataOperationImage get data operation image for cache runtime by using worker image.
func (e *CacheEngine) getDataOperationImage(runtime *v1alpha1.CacheRuntime, runtimeClass *v1alpha1.CacheRuntimeClass) (image string, err error) {
	// runtime definition has higher priority than runtime class
	if runtimeClass.Topology.Worker != nil && len(runtimeClass.Topology.Worker.Template.Spec.Containers) > 0 {
		// container [0] is the cache engine image
		image = runtimeClass.Topology.Worker.Template.Spec.Containers[0].Image
	}
	if len(runtime.Spec.Worker.RuntimeVersion.Image) > 0 && len(runtime.Spec.Worker.RuntimeVersion.ImageTag) > 0 {
		image = runtime.Spec.Worker.RuntimeVersion.Image + ":" + runtime.Spec.Worker.RuntimeVersion.ImageTag
	}

	if len(image) == 0 {
		return "", fmt.Errorf("no image for runtime, name: %s, namespace: %s", runtime.Name, runtime.Namespace)
	}

	return image, nil
}
