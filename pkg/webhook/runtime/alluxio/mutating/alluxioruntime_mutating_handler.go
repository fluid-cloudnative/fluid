package mutating

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
	"time"
)

var _ common.AdmissionHandler = &UpdateAlluxioRuntimeHandler{}
var _ admission.DecoderInjector = &UpdateAlluxioRuntimeHandler{}

const hostPathAnnoKey = "volume.fluid.io/hostpath"

// definitions for volume attribute
const (
	volumeMediumTypeKey1 = "diskType"
	volumeMediumTypeKey2 = "poolClass"
)

type UpdateAlluxioRuntimeHandler struct {
	Client client.Client
	// A decoder will be automatically injected
	decoder *admission.Decoder
}

func (a *UpdateAlluxioRuntimeHandler) InjectDecoder(decoder *admission.Decoder) error {
	a.decoder = decoder
	return nil
}

func (a *UpdateAlluxioRuntimeHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	defer utils.TimeTrack(time.Now(), "CreateUpdatePodForSchedulingHandler.Handle",
		"req.name", req.Name, "req.namespace", req.Namespace)

	if utils.GetBoolValueFromEnv(common.EnvDisableInjection, false) {
		return admission.Allowed("skip mutating the alluxio runtime because global injection is disabled")
	}

	var setupLog = ctrl.Log.WithName("handle")
	alluxioRuntime := &v1alpha1.AlluxioRuntime{}
	err := a.decoder.Decode(req, alluxioRuntime)
	if err != nil {
		setupLog.Error(err, "unable to decoder alluxioruntime from req")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check if alluxio runtime is needed mutate
	if needMutate, err := a.needMutate(alluxioRuntime); err != nil {
		setupLog.Error(err, "unable to judge if alluxio runtime needed mutate")
		return admission.Errored(http.StatusInternalServerError, err)
	} else if !needMutate {
		return admission.Allowed("skip mutating the alluxio runtime because no need to mutate")
	}

	// mutate alluxioruntime
	if err = a.doMutate(alluxioRuntime); err != nil {
		setupLog.Error(err, "failed to mutate alluxio runtime")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// construct mutate result
	marshaledRuntime, err := json.Marshal(alluxioRuntime)
	if err != nil {
		setupLog.Error(err, "unable to marshal alluxioruntime")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	resp := admission.PatchResponseFromRaw(req.Object.Raw, marshaledRuntime)
	setupLog.V(1).Info("patch response", "name", alluxioRuntime.GetName(), "patches", utils.DumpJSON(resp.Patch))
	return resp
}

func (a *UpdateAlluxioRuntimeHandler) Setup(client client.Client) {
	a.Client = client
}

// needMutate returns true if pvc specified in runtime configuration
func (a *UpdateAlluxioRuntimeHandler) needMutate(ar *v1alpha1.AlluxioRuntime) (bool, error) {
	for _, levelStore := range ar.Spec.TieredStore.Levels {
		if levelStore.PersistentVolumeClaim != "" {
			return true, nil
		}
	}
	return false, nil
}

func (a *UpdateAlluxioRuntimeHandler) doMutate(ar *v1alpha1.AlluxioRuntime) error {
	var setupLog = ctrl.Log.WithName("doMutate")
	levelStoresNeedMutate := map[int]*v1alpha1.Level{}
	for i, levelStore := range ar.Spec.TieredStore.Levels {
		if levelStore.PersistentVolumeClaim != "" {
			levelStoresNeedMutate[i] = levelStore.DeepCopy()
			setupLog.Info("found pvc in level store", "pvc", levelStore.PersistentVolumeClaim)
		}
	}

	setupLog.Info("finish found pvc", "total", len(levelStoresNeedMutate))

	for i, levelStore := range levelStoresNeedMutate {
		newLevelStore, err := a.reconstructLevelStore(levelStore)
		if err != nil {
			setupLog.Error(err, "failed to reconstruct level store")
			return err
		}
		ar.Spec.TieredStore.Levels[i] = *newLevelStore
	}

	return nil
}

func (a *UpdateAlluxioRuntimeHandler) reconstructLevelStore(levelStore *v1alpha1.Level) (*v1alpha1.Level, error) {
	var setupLog = ctrl.Log.WithName("reconstructLevelStore")
	pvcNamespacedName := levelStore.PersistentVolumeClaim
	ss := strings.Split(pvcNamespacedName, "/")
	if len(ss) != 2 {
		return nil, fmt.Errorf("%s is invalid, specify it in <namespace>/<name> format", pvcNamespacedName)
	}

	pvc := v12.PersistentVolumeClaim{}
	pvc.Namespace = ss[0]
	pvc.Name = ss[1]
	if err := a.Client.Get(context.Background(), client.ObjectKey{Namespace: pvc.Namespace, Name: pvc.Name}, &pvc); err != nil {
		return nil, err
	}

	if pvc.Spec.StorageClassName == nil {
		setupLog.Info("skip mutate level store because of no storageclass found", "pvc", pvcNamespacedName)
		return levelStore, nil
	}

	// fetch the storageclass
	sc := v1.StorageClass{}
	if err := a.Client.Get(context.Background(), client.ObjectKey{Name: *pvc.Spec.StorageClassName}, &sc); err != nil {
		return nil, err
	}

	if !strings.HasSuffix(sc.Provisioner, "hwameistor.io") {
		setupLog.Info("skip mutate level store because of using unknown storageclass",
			"pvc", pvcNamespacedName, "storageclass", pvc.Spec.StorageClassName)
		return levelStore, nil
	}

	// fetch the persistentvolume
	pv := v12.PersistentVolume{}
	if err := a.Client.Get(context.Background(), client.ObjectKey{Name: pvc.Spec.VolumeName}, &pv); err != nil {
		return nil, err
	}

	if pv.Annotations == nil {
		setupLog.Info("skip mutate level store because of mountpoint not found in persistentvolume",
			"pvc", pvcNamespacedName, "storageclass", pvc.Spec.StorageClassName, "pv", pv.Name)
		return levelStore, nil
	}

	volumeHostPath := pv.Annotations[hostPathAnnoKey]
	newLevelStore := levelStore.DeepCopy()
	newLevelStore.Path = volumeHostPath
	newLevelStore.MediumType = getMediumType(sc.Parameters)

	setupLog.Info("reconstructLevelStore finished", "Path", newLevelStore.Path, "mediumType", newLevelStore.MediumType)

	return newLevelStore, nil
}

func getMediumType(m map[string]string) common.MediumType {
	if v, ok := m[volumeMediumTypeKey1]; ok {
		return common.MediumType(v)
	}
	return common.MediumType(m[volumeMediumTypeKey2])
}
