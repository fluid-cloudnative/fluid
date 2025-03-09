package componenthelper

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	appsapplyv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/componenthelper/utils"
)

type DaemonsetReconciler struct {
	scheme *runtime.Scheme
	client client.Client
}

var _ ComponentHelper = &DaemonsetReconciler{}

func NewDaemonsetReconciler(scheme *runtime.Scheme, client client.Client) *DaemonsetReconciler {
	return &DaemonsetReconciler{scheme: scheme, client: client}
}

func (r *DaemonsetReconciler) GetComponentTopologyInfo(ctx context.Context, component *common.CacheRuntimeComponentValue) (common.TopologyConfig, error) {
	workload := &appsv1.DaemonSet{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, workload); err != nil {
		return common.TopologyConfig{}, err
	}

	podList := &corev1.PodList{}
	if err := r.client.List(context.TODO(), podList, client.InNamespace(component.Namespace), client.MatchingLabels(workload.Spec.Selector.MatchLabels)); err != nil {
		return common.TopologyConfig{}, err
	}

	topologyComponent := common.TopologyConfig{}
	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			continue
		}
		podConfig := common.PodConfig{
			PodName: pod.Name,
			PodIP:   pod.Status.PodIP,
		}
		for _, port := range pod.Spec.Containers[0].Ports {
			podConfig.Ports = append(podConfig.Ports, common.PortConfig{
				Name: port.Name,
				Port: port.ContainerPort,
			})
		}
		topologyComponent.PodConfigs = append(topologyComponent.PodConfigs, podConfig)
	}

	if component.Service != nil {
		svc := &corev1.Service{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, svc)
		if err != nil {
			return common.TopologyConfig{}, err
		}

		topologyComponent.Service = common.CacheRuntimeComponentServiceConfig{
			Name: svc.Name,
		}
	}

	return topologyComponent, nil
}

func (r *DaemonsetReconciler) Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if err := r.reconcileDaemonset(ctx, component); err != nil {
		return err
	}

	return r.reconcileService(ctx, component)
}

func (r *DaemonsetReconciler) reconcileDaemonset(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("start to reconciling sts workload")

	oldDs := &appsv1.DaemonSet{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, oldDs); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	dsApplyConfig, err := r.constructDaemonSetApplyConfiguration(ctx, component, oldDs)
	if err != nil {
		logger.Error(err, "Failed to construct statefulset apply configuration")
		return err
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dsApplyConfig)
	if err != nil {
		logger.Error(err, "Converting obj apply configuration to json.")
		return err
	}

	newDs := &appsv1.DaemonSet{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj, newDs); err != nil {
		return fmt.Errorf("convert stsApplyConfig to sts error: %s", err.Error())
	}

	equal, err := SemanticallyEqualDaemonset(newDs, newDs)
	if equal {
		logger.V(1).Info("sts equal, skip reconcile")
		return nil
	}

	logger.V(1).Info(fmt.Sprintf("sts not equal, diff: %s", err.Error()))

	if err := utils.PatchObjectApplyConfiguration(ctx, r.client, dsApplyConfig, utils.PatchSpec); err != nil {
		logger.Error(err, "Failed to patch statefulset apply configuration")
		return err
	}

	return nil
}

func (r *DaemonsetReconciler) constructDaemonSetApplyConfiguration(
	ctx context.Context,
	component *common.CacheRuntimeComponentValue,
	oldDs *appsv1.DaemonSet,
) (*appsapplyv1.DaemonSetApplyConfiguration, error) {
	matchLabels := common.GetCommonLabelsFromComponent(component)
	if oldDs.UID != "" {
		// do not update selector when workload exists
		matchLabels = oldDs.Spec.Selector.MatchLabels
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&component.PodTemplateSpec)
	if err != nil {
		return nil, err
	}
	var podTemplateApplyConfiguration *coreapplyv1.PodTemplateSpecApplyConfiguration
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &podTemplateApplyConfiguration)
	if err != nil {
		return nil, err
	}
	podTemplateApplyConfiguration.WithLabels(matchLabels)

	// construct daemonset apply configuration
	daemonsetSetConfig := appsapplyv1.DaemonSet(component.Name, component.Namespace).
		WithSpec(appsapplyv1.DaemonSetSpec().
			WithTemplate(podTemplateApplyConfiguration).
			WithSelector(metaapplyv1.LabelSelector().
				WithMatchLabels(matchLabels))).
		WithLabels(matchLabels).
		WithOwnerReferences(metaapplyv1.OwnerReference().
			WithAPIVersion(component.Owner.APIVersion).
			WithKind(component.Owner.Kind).
			WithName(component.Owner.Name).
			WithUID(types.UID(component.Owner.UID)).
			WithBlockOwnerDeletion(true).
			WithController(true),
		)
	return daemonsetSetConfig, nil
}

func (r *DaemonsetReconciler) CheckComponentExist(ctx context.Context, component *common.CacheRuntimeComponentValue) (bool, error) {
	ds := &appsv1.DaemonSet{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DaemonsetReconciler) ConstructComponentStatus(ctx context.Context, component *common.CacheRuntimeComponentValue) (datav1alpha1.RuntimeComponentStatus, error) {
	ds := &appsv1.DaemonSet{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds); err != nil {
		return datav1alpha1.RuntimeComponentStatus{}, err
	}

	return datav1alpha1.RuntimeComponentStatus{
		DesiredReplicas:     ds.Status.DesiredNumberScheduled,
		CurrentReplicas:     ds.Status.CurrentNumberScheduled,
		AvailableReplicas:   ds.Status.NumberAvailable,
		UnavailableReplicas: ds.Status.NumberUnavailable,
		ReadyReplicas:       ds.Status.NumberReady,
	}, nil
}

func (r *DaemonsetReconciler) CheckComponentReady(ctx context.Context, component *common.CacheRuntimeComponentValue) (bool, error) {
	ds := &appsv1.DaemonSet{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds); err != nil {
		return false, err
	}
	return ds.Status.NumberUnavailable == 0, nil
}

func (r *DaemonsetReconciler) CleanupOrphanedComponentResources(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	return nil
}

func (r *DaemonsetReconciler) reconcileService(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if component == nil || component.Service == nil {
		return nil
	}

	logger := log.FromContext(ctx)
	logger.Info("start to reconciling headless service")

	sts := &appsv1.StatefulSet{}
	err := r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("get sts error, skip reconcile svc. error:  %s", err.Error())
	}

	svcApplyConfig, err := r.constructServiceApplyConfiguration(ctx, component, sts)
	if err != nil {
		return err
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svcApplyConfig)
	if err != nil {
		logger.Error(err, "Converting obj apply configuration to json.")
		return err
	}

	newSvc := &corev1.Service{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj, newSvc); err != nil {
		return fmt.Errorf("convert svcApplyConfig to svc error: %s", err.Error())
	}

	oldSvc := &corev1.Service{}
	err = r.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, oldSvc)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	equal, err := SemanticallyEqualService(oldSvc, newSvc)
	if equal {
		logger.V(1).Info("svc equal, skip reconcile")
		return nil
	}

	logger.V(1).Info(fmt.Sprintf("svc not equal, diff: %s", err.Error()))

	if err := utils.PatchObjectApplyConfiguration(ctx, r.client, svcApplyConfig, utils.PatchSpec); err != nil {
		logger.Error(err, "Failed to patch svc apply configuration")
		return err
	}

	return nil
}

func (r *DaemonsetReconciler) constructServiceApplyConfiguration(
	ctx context.Context,
	component *common.CacheRuntimeComponentValue,
	sts *appsv1.StatefulSet,
) (*coreapplyv1.ServiceApplyConfiguration, error) {
	matchLabels := common.GetCommonLabelsFromComponent(component)
	if sts.UID != "" {
		// do not update selector when workload exists
		matchLabels = sts.Spec.Selector.MatchLabels
	}
	serviceConfig := coreapplyv1.Service(component.Name, component.Namespace).
		WithSpec(coreapplyv1.ServiceSpec().
			WithClusterIP("None").
			WithSelector(matchLabels).
			WithPublishNotReadyAddresses(true)).
		WithLabels(matchLabels).
		WithOwnerReferences(metaapplyv1.OwnerReference().
			WithAPIVersion(sts.APIVersion).
			WithKind(sts.Kind).
			WithName(sts.Name).
			WithUID(sts.GetUID()).
			WithBlockOwnerDeletion(true),
		)
	return serviceConfig, nil
}

func SemanticallyEqualDaemonset(oldDs, newDs *appsv1.DaemonSet) (bool, error) {
	if oldDs == nil || oldDs.UID == "" {
		return false, errors.New("old sts not exist")
	}
	if newDs == nil {
		return false, fmt.Errorf("new sts is nil")
	}

	if equal, err := objectMetaEqual(oldDs.ObjectMeta, newDs.ObjectMeta); !equal {
		return false, fmt.Errorf("objectMeta not equal: %s", err.Error())
	}

	if equal, err := daemonsetSpecEqual(oldDs.Spec, newDs.Spec); !equal {
		return false, fmt.Errorf("spec not equal: %s", err.Error())
	}
	return true, nil
}

func daemonsetSpecEqual(spec1, spec2 appsv1.DaemonSetSpec) (bool, error) {
	if !reflect.DeepEqual(spec1.Selector, spec2.Selector) {
		return false, fmt.Errorf("selector not equal, old: %v, new: %v", spec1.Selector, spec2.Selector)
	}

	if equal, err := podTemplateSpecEqual(spec1.Template, spec2.Template); !equal {
		return false, fmt.Errorf("podTemplateSpec not equal, %s", err.Error())
	}

	return true, nil
}
