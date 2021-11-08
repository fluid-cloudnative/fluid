package ctrl

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The common part of the engine which can be reused
type Helper struct {
	runtime base.RuntimeInfoInterface

	client client.Client
}

func BuildHelper(runtime base.RuntimeInfoInterface, client client.Client) *Helper {
	return &Helper{
		runtime: runtime,
		client:  client,
	}
}

// BuildWorkersAffinity builds workers affinity if it doesn't have
func (e *Helper) BuildWorkersAffinity(workers *appsv1.StatefulSet) (workersToUpdate *appsv1.StatefulSet, err error) {
	// TODO: for now, runtime affinity can't be set by user, so we can assume the affinity is nil in the first time.
	// We need to enhance it in future
	workersToUpdate = workers.DeepCopy()
	var (
		name      = e.runtime.GetName()
		namespace = e.runtime.GetNamespace()
	)

	if workersToUpdate.Spec.Template.Spec.Affinity == nil {
		workersToUpdate.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		dataset, err := utils.GetDataset(e.client, name, namespace)
		if err != nil {
			return workersToUpdate, err
		}
		// 1. Set pod anti affinity(required) for same dataset (Current using port conflict for scheduling, no need to do)

		// 2. Set pod anti affinity for the different dataset
		if dataset.IsExclusiveMode() {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset",
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		} else {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						// The default weight is 50
						Weight: 50,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "fluid.io/dataset",
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset-placement",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{string(datav1alpha1.ExclusiveMode)},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		}

		// 3. Perefer to locate on the node which already has fuse on it
		if workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity == nil {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
		}

		if len(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
		}

		workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
			append(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				corev1.PreferredSchedulingTerm{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      e.runtime.GetFuseLabelName(),
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},
						},
					},
				})

		// 3. set node affinity if possible
		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
					dataset.Spec.NodeAffinity.Required
			}
		}

		err = e.client.Update(context.TODO(), workersToUpdate)
		if err != nil {
			return workersToUpdate, err
		}

	}

	return
}
