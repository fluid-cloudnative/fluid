/*
Copyright 2026 The Fluid Authors.
Copyright 2019 The Kruise Authors.
Copyright 2016 The Kubernetes Authors.

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

package advancedstatefulset

import (
	"context"
	"flag"
	"fmt"
	"time"

	history "github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/controllerhistory"
	utildiscovery "github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/discovery"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/expectations"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/inplaceupdate"
	kubecontroller2 "github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/kubecontroller"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/ratelimiter"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/requeueduration"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/revisionadapter"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/util"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	discoverylib "k8s.io/client-go/discovery"
	kubeclientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
)

func init() {
	flag.IntVar(&concurrentReconciles, "advancedstatefulset-workers", concurrentReconciles, "Max concurrent workers for AdvancedStatefulSet controller.")
}

var (
	// controllerKind contains the schema.GroupVersionKind for this controller type.
	controllerKind       = workloadv1alpha1.SchemeGroupVersion.WithKind("AdvancedStatefulSet")
	concurrentReconciles = 3

	updateExpectations = expectations.NewUpdateExpectations(revisionadapter.NewDefaultImpl())
	// this is a short cut for any sub-functions to notify the reconcile how long to wait to requeue
	durationStore = requeueduration.DurationStore{}
)

// Add creates a new AdvancedStatefulSet Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	// Initialize discovery client for GVK detection
	dc, err := discoverylib.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	utildiscovery.Init(dc)

	if !utildiscovery.DiscoverGVK(controllerKind) {
		return nil
	}
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	cacher := mgr.GetCache()
	podInformer, err := cacher.GetInformerForKind(context.TODO(), v1.SchemeGroupVersion.WithKind("Pod"))
	if err != nil {
		return nil, err
	}
	pvcInformer, err := cacher.GetInformerForKind(context.TODO(), v1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))
	if err != nil {
		return nil, err
	}
	revInformer, err := cacher.GetInformerForKind(context.TODO(), appsv1.SchemeGroupVersion.WithKind("ControllerRevision"))
	if err != nil {
		return nil, err
	}

	podLister := corelisters.NewPodLister(podInformer.(toolscache.SharedIndexInformer).GetIndexer())
	pvcLister := corelisters.NewPersistentVolumeClaimLister(pvcInformer.(toolscache.SharedIndexInformer).GetIndexer())

	genericCfg := rest.CopyConfig(mgr.GetConfig())
	genericCfg.UserAgent = "advancedstatefulset-controller"
	kubeClient, err := kubeclientset.NewForConfig(genericCfg)
	if err != nil {
		return nil, err
	}
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(workloadv1alpha1.Scheme, v1.EventSource{Component: "advancedstatefulset-controller"})

	runtimeClient := mgr.GetClient()
	return &ReconcileStatefulSet{
		Client: runtimeClient,
		control: NewDefaultStatefulSetControl(
			NewStatefulPodControl(
				kubeClient,
				podLister,
				pvcLister,
				recorder),
			inplaceupdate.New(runtimeClient, revisionadapter.NewDefaultImpl()),
			lifecycle.New(runtimeClient),
			NewRealStatefulSetStatusUpdater(runtimeClient),
			history.NewHistory(kubeClient, appslisters.NewControllerRevisionLister(revInformer.(toolscache.SharedIndexInformer).GetIndexer())),
			recorder,
		),
		podControl: kubecontroller2.RealPodControl{KubeClient: kubeClient, Recorder: recorder},
		podLister:  podLister,
	}, nil
}

var _ reconcile.Reconciler = &ReconcileStatefulSet{}

// ReconcileStatefulSet reconciles a AdvancedStatefulSet object
type ReconcileStatefulSet struct {
	// client provides controller-runtime client for AdvancedStatefulSet operations
	runtimeclient.Client
	// control returns an interface capable of syncing a stateful set.
	// Abstracted out for testing.
	control StatefulSetControlInterface
	// podControl is used for patching pods.
	podControl kubecontroller2.PodControlInterface
	// podLister is able to list/get pods from a shared informer's store
	podLister corelisters.PodLister
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("advancedstatefulset-controller", mgr, controller.Options{
		Reconciler: r, MaxConcurrentReconciles: concurrentReconciles, CacheSyncTimeout: util.GetControllerCacheSyncTimeout(),
		RateLimiter: ratelimiter.DefaultControllerRateLimiter[reconcile.Request]()})
	if err != nil {
		return err
	}

	// Watch for changes to AdvancedStatefulSet
	err = c.Watch(source.Kind(mgr.GetCache(), &workloadv1alpha1.AdvancedStatefulSet{}),
		&handler.EnqueueRequestForObject{},
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldSS, ok1 := e.ObjectOld.(*workloadv1alpha1.AdvancedStatefulSet)
				newSS, ok2 := e.ObjectNew.(*workloadv1alpha1.AdvancedStatefulSet)
				if ok1 && ok2 && oldSS.Status.Replicas != newSS.Status.Replicas {
					klog.V(4).InfoS("Observed updated replica count for AdvancedStatefulSet",
						"statefulSet", klog.KObj(newSS), "oldReplicas", oldSS.Status.Replicas, "newReplicas", newSS.Status.Replicas)
				}
				return true
			},
		})
	if err != nil {
		return err
	}

	// Watch for changes to PVC patched by AdvancedStatefulSet
	err = c.Watch(source.Kind(mgr.GetCache(), &v1.PersistentVolumeClaim{}), &pvcEventHandler{})
	if err != nil {
		return err
	}

	// Watch for changes to Pod created by AdvancedStatefulSet
	err = c.Watch(source.Kind(mgr.GetCache(), &v1.Pod{}),
		handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &workloadv1alpha1.AdvancedStatefulSet{}, handler.OnlyControllerOwner()))
	if err != nil {
		return err
	}

	klog.V(4).InfoS("Finished to add advancedstatefulset-controller")

	return nil
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=controllerrevisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workload.fluid.io,resources=advancedstatefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workload.fluid.io,resources=advancedstatefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workload.fluid.io,resources=advancedstatefulsets/finalizers,verbs=update

// Reconcile reads that state of the cluster for a AdvancedStatefulSet object and makes changes based on the state read
// and what is in the AdvancedStatefulSet.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Pods
func (ssc *ReconcileStatefulSet) Reconcile(ctx context.Context, request reconcile.Request) (res reconcile.Result, retErr error) {
	key := request.NamespacedName.String()
	namespace := request.Namespace
	name := request.Name

	startTime := time.Now()
	defer func() {
		if retErr == nil {
			if res.Requeue || res.RequeueAfter > 0 {
				klog.InfoS("Finished syncing AdvancedStatefulSet", "statefulSet", request, "elapsedTime", time.Since(startTime), "result", res)
			} else {
				klog.InfoS("Finished syncing AdvancedStatefulSet", "statefulSet", request, "elapsedTime", time.Since(startTime))
			}
		} else {
			klog.ErrorS(retErr, "Finished syncing AdvancedStatefulSet error", "statefulSet", request, "elapsedTime", time.Since(startTime))
		}
	}()

	set := &workloadv1alpha1.AdvancedStatefulSet{}
	if err := ssc.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, set); err != nil {
		if errors.IsNotFound(err) {
			klog.InfoS("AdvancedStatefulSet deleted", "statefulSet", key)
			updateExpectations.DeleteExpectations(key)
			return reconcile.Result{}, nil
		}
		utilruntime.HandleError(fmt.Errorf("unable to retrieve AdvancedStatefulSet %v from store: %v", key, err))
		return reconcile.Result{}, err
	}

	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting AdvancedStatefulSet %v selector: %v", key, err))
		// This is a non-transient error, so don't retry.
		return reconcile.Result{}, nil
	}

	if err := ssc.adoptOrphanRevisions(set); err != nil {
		return reconcile.Result{}, err
	}

	pods, err := ssc.getPodsForStatefulSet(ctx, set, selector)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = ssc.syncStatefulSet(ctx, set, pods)
	return reconcile.Result{RequeueAfter: durationStore.Pop(getStatefulSetKey(set))}, err
}

// adoptOrphanRevisions adopts any orphaned ControllerRevisions matched by set's Selector.
func (ssc *ReconcileStatefulSet) adoptOrphanRevisions(set *workloadv1alpha1.AdvancedStatefulSet) error {
	revisions, err := ssc.control.ListRevisions(set)
	if err != nil {
		return err
	}
	orphanRevisions := make([]*appsv1.ControllerRevision, 0)
	for i := range revisions {
		if metav1.GetControllerOf(revisions[i]) == nil {
			orphanRevisions = append(orphanRevisions, revisions[i])
		}
	}
	if len(orphanRevisions) > 0 {
		fresh := &workloadv1alpha1.AdvancedStatefulSet{}
		if err := ssc.Get(context.TODO(), types.NamespacedName{Namespace: set.Namespace, Name: set.Name}, fresh); err != nil {
			return err
		}
		if fresh.UID != set.UID {
			return fmt.Errorf("original AdvancedStatefulSet %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return ssc.control.AdoptOrphanRevisions(set, orphanRevisions)
	}
	return nil
}

// getPodsForStatefulSet returns the Pods that a given AdvancedStatefulSet should manage.
// It also reconciles ControllerRef by adopting/orphaning.
//
// NOTE: Returned Pods are pointers to objects from the cache.
//
//	If you need to modify one, you need to copy it first.
func (ssc *ReconcileStatefulSet) getPodsForStatefulSet(ctx context.Context, set *workloadv1alpha1.AdvancedStatefulSet, selector labels.Selector) ([]*v1.Pod, error) {
	// List all pods to include the pods that don't match the selector anymore but
	// has a ControllerRef pointing to this AdvancedStatefulSet.
	pods, err := ssc.podLister.Pods(set.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	filter := func(pod *v1.Pod) bool {
		// Only claim if it matches our AdvancedStatefulSet name. Otherwise release/ignore.
		return isMemberOf(set, pod)
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := kubecontroller2.RecheckDeletionTimestamp(func(ctx context.Context) (metav1.Object, error) {
		fresh := &workloadv1alpha1.AdvancedStatefulSet{}
		if err := ssc.Get(ctx, types.NamespacedName{Namespace: set.Namespace, Name: set.Name}, fresh); err != nil {
			return nil, err
		}
		if fresh.UID != set.UID {
			return nil, fmt.Errorf("original AdvancedStatefulSet %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return fresh, nil
	})

	cm := kubecontroller2.NewPodControllerRefManager(ssc.podControl, set, selector, controllerKind, canAdoptFunc)
	return cm.ClaimPods(ctx, pods, filter)
}

// syncStatefulSet syncs a tuple of (advancedstatefulset, []*v1.Pod).
func (ssc *ReconcileStatefulSet) syncStatefulSet(ctx context.Context, set *workloadv1alpha1.AdvancedStatefulSet, pods []*v1.Pod) error {
	klog.V(4).InfoS("Syncing AdvancedStatefulSet with pods", "statefulSet", klog.KObj(set), "podCount", len(pods))
	// TODO: investigate where we mutate the set during the update as it is not obvious.
	if err := ssc.control.UpdateStatefulSet(ctx, set.DeepCopy(), pods); err != nil {
		return err
	}
	klog.V(4).InfoS("Successfully synced AdvancedStatefulSet", "statefulSet", klog.KObj(set))
	return nil
}
