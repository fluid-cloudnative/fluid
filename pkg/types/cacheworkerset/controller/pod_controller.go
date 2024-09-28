package statefulset

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	apps "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/client/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DataLoadReconciler reconciles a DataLoad object
type AstsReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *AstsReconciler) ControllerName() string {
	return "advancedstatefulset"
}

// Reconcile 实现了 Reconciler 接口
func (r *AstsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the StatefulSet instance
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, req.NamespacedName, statefulSet)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Add your reconciliation logic here...
	// 例如，获取相关的 Pods 并进行同步操作

	// 示例逻辑
	pods, err := r.getPodsForStatefulSet(statefulSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 同步 StatefulSet 和 Pods
	err = r.syncStatefulSet(statefulSet, pods)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager 配置 Controller 与 Manager 的交互
func (r *AstsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.findStatefulSetForPod),
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return true
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
			},
		).
		Complete(r)
}

// findStatefulSetForPod 找到 Pod 对应的 StatefulSet
func (r *AstsReconciler) findStatefulSetForPod(o handler.MapObject) []reconcile.Request {
	pod := o.(*corev1.Pod)
	if ownerRef := metav1.GetControllerOf(pod); ownerRef != nil {
		namespace := pod.Namespace
		name := ownerRef.Name
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}}
	}
	return nil
}

// getPodsForStatefulSet 获取 StatefulSet 对应的 Pods
func (r *AstsReconciler) getPodsForStatefulSet(set *appsv1.StatefulSet) ([]*corev1.Pod, error) {
	pods := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(set.Namespace),
		//增加或者antion 标签过滤	nota
		client.MatchingLabelsSelector{Selector: metav1.LabelSelectorAsSelector(set.Spec.Selector)},
	}
	if err := r.List(context.TODO(), pods, listOpts...); err != nil {
		return nil, err
	}

	filter := func(pod *corev1.Pod) bool {
		return isMemberOf(set, pod)
	}

	return filterPods(pods.Items, filter), nil
}

// filterPods 过滤 Pods
func filterPods(pods []corev1.Pod, filter func(*corev1.Pod) bool) []*corev1.Pod {
	var result []*corev1.Pod
	for _, pod := range pods {
		if filter(&pod) {
			result = append(result, &pod)
		}
	}
	return result
}

// isMemberOf 检查 Pod 是否属于 StatefulSet
func isMemberOf(set *appsv1.StatefulSet, pod *corev1.Pod) bool {
	// 示例逻辑，具体实现根据需求调整
	return true
}

// syncStatefulSet 同步 StatefulSet 和 Pods
func (r *AstsReconciler) syncAdvancedStatefulSet(set *appsv1.StatefulSet, pods []*corev1.Pod) error {
	// 示例逻辑
	klog.V(4).Infof("Syncing StatefulSet %v/%v with %d pods", set.Namespace, set.Name, len(pods))

	// 具体同步逻辑
	return nil
}

// main 函数
func AstsControllerWorker() {
	var (
		scheme = runtime.NewScheme()
	)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&AstsReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("StatefulSet"),
		Scheme: scheme,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "Controller", "StatefulSet")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// syncStatefulSet syncs a tuple of (statefulset, []*v1.Pod).
func (ssc *AstsReconciler) syncStatefulSet(set *apps.StatefulSet, pods []*v1.Pod) error {
	klog.V(4).Infof("Syncing StatefulSet %v/%v with %d pods", set.Namespace, set.Name, len(pods))
	// TODO: investigate where we mutate the set during the update as it is not obvious.
	if err := ssc.control.UpdateStatefulSet(set.DeepCopy(), pods); err != nil {
		return err
	}
	klog.V(4).Infof("Successfully synced StatefulSet %s/%s successful", set.Namespace, set.Name)
	return nil
}
