package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	FieldManager = "rbg"

	PatchAll    PatchType = "all"
	PatchSpec   PatchType = "spec"
	PatchStatus PatchType = "status"
)

type PatchType string

func PatchObjectApplyConfiguration(ctx context.Context, k8sClient client.Client, objApplyConfig interface{}, patchType PatchType) error {
	logger := log.FromContext(ctx)
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objApplyConfig)
	if err != nil {
		logger.Error(err, "Converting obj apply configuration to json.")
		return err
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	logger.V(1).Info("patch content", "patchObject", patch.Object)

	// Use server side apply and add fieldmanager to the rbg owned fields
	// If there are conflicts in the fields owned by the rbg controller, rbg will obtain the ownership and force override
	// these fields to the ones desired by the rbg controller
	// TODO b/316776287 add E2E test for SSA
	if patchType == PatchSpec || patchType == PatchAll {
		err = k8sClient.Patch(ctx, patch, client.Apply, &client.PatchOptions{
			FieldManager: FieldManager,
			Force:        ptr.To[bool](true),
		})
		if err != nil {
			logger.Error(err, "Using server side apply to patch object")
			return err
		}
	}

	if patchType == PatchStatus || patchType == PatchAll {
		err = k8sClient.Status().Patch(ctx, patch, client.Apply,
			&client.SubResourcePatchOptions{
				PatchOptions: client.PatchOptions{
					FieldManager: FieldManager,
					Force:        ptr.To[bool](true),
				},
			})
		if err != nil {
			logger.Error(err, "Using server side apply to patch object status")
			return err
		}
	}

	return nil
}

func ContainsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func PrettyJson(object interface{}) string {
	b, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		fmt.Printf("ERROR: PrettyJson, %v\n %s\n", err, b)
		return ""
	}
	return string(b)
}

func FilterSystemAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	filtered := make(map[string]string)
	for k, v := range annotations {
		if !strings.HasPrefix(k, "deployment.kubernetes.io/revision") &&
			!strings.HasPrefix(k, "rolebasedgroup.workloads.x-k8s.io/") &&
			!strings.HasPrefix(k, "app.kubernetes.io/") {
			filtered[k] = v
		}
	}
	return filtered
}

// DumpJSON returns the JSON encoding
func DumpJSON(o interface{}) string {
	j, _ := json.Marshal(o)
	return string(j)
}

func NonZeroValue(value int32) int32 {
	if value < 0 {
		return 0
	}
	return value
}

func FilterSystemEnvs(envs []corev1.EnvVar) []corev1.EnvVar {
	var filtered []corev1.EnvVar
	for _, env := range envs {
		if !strings.HasPrefix(env.Name, "FLUID_") && env.Name != "RBG_GROUP_SIZE" {
			filtered = append(filtered, env)
		}
	}
	return filtered
}

func FilterSystemLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}

	filtered := make(map[string]string)
	for k, v := range labels {

		if !strings.Contains(k, common.LabelAnnotationPrefix) {
			filtered[k] = v
		}
	}
	return filtered

}
