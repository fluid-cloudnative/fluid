package referencedataset

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	syncRetryDurationEnv     string = "FLUID_SYNC_RETRY_DURATION"
	defaultSyncRetryDuration        = 5 * time.Second
)

// Use compiler to check if the struct implements all the interface
var _ base.Engine = (*ReferenceDatasetEngine)(nil)

// ReferenceDatasetEngine is used for handling datasets mounting another dataset
type ReferenceDatasetEngine struct {
	Id string
	client.Client
	Log logr.Logger

	name      string
	namespace string

	syncRetryDuration time.Duration
	timeOfLastSync    time.Time

	runtimeType string
	// use getRuntimeInfo instead of directly use this field
	runtimeInfo base.RuntimeInfoInterface
	// mounted dataset corresponding runtimeInfo,  use getMountedRuntimeInfo instead of directly use this field
	mountedRuntimeInfo base.RuntimeInfoInterface
}

// ID returns the id of the engine
func (e *ReferenceDatasetEngine) ID() string {
	return e.Id
}

// BuildReferenceDatasetThinEngine build engine for handling virtual dataset
func BuildReferenceDatasetThinEngine(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	// currently, ctx has no dataset
	engine := &ReferenceDatasetEngine{
		Id:          id,
		Client:      ctx.Client,
		name:        ctx.Name,
		namespace:   ctx.Namespace,
		runtimeType: ctx.RuntimeType,
	}
	engine.Log = ctx.Log.WithValues("virtual engine", ctx.RuntimeType).WithValues("id", id)

	// check if support the dataset mount format
	err := engine.checkDatasetMountSupport()
	if err != nil {
		return nil, err
	}

	// Build and setup runtime info
	_, err = engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	// set sync duration
	duration, err := getSyncRetryDuration()
	if err != nil {
		engine.Log.Error(err, "Failed to parse syncRetryDurationEnv: FLUID_SYNC_RETRY_DURATION, use the default setting")
	}
	if duration != nil {
		engine.syncRetryDuration = *duration
	} else {
		engine.syncRetryDuration = defaultSyncRetryDuration
	}
	engine.timeOfLastSync = time.Now().Add(-engine.syncRetryDuration)
	engine.Log.Info("Set the syncRetryDuration", "syncRetryDuration", engine.syncRetryDuration)

	// get the mountedRuntimeInfo
	_, err = engine.getMountedRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get mounted dataset's runtime info", ctx.Name)
	}

	return engine, err
}

func (e *ReferenceDatasetEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	// 1. get the physical datasets according the virtual dataset
	dataset := ctx.Dataset

	physicalNameSpacedNames := getMountedDatasetNamespacedName(dataset)
	if len(physicalNameSpacedNames) != 1 {
		return false, fmt.Errorf("ThinEngine with no profile name can only handle dataset only mounting one dataset")
	}
	namespacedName := physicalNameSpacedNames[0]

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		mountedDataset, err := utils.GetDataset(ctx.Client, namespacedName.Name, namespacedName.Namespace)
		if err != nil {
			return err
		}

		// 2. get runtime according to dataset status runtimes field
		runtimes := mountedDataset.Status.Runtimes
		if len(runtimes) == 0 {
			return fmt.Errorf("mounting dataset is not bound to a runtime yet")
		}

		// 3. add this dataset to mounted dataset DatasetRef field
		datasetRefName := getDatasetRefName(dataset.Name, dataset.Namespace)
		if !utils.ContainsString(mountedDataset.Status.DatasetRef, datasetRefName) {
			newDataset := mountedDataset.DeepCopy()
			newDataset.Status.DatasetRef = append(newDataset.Status.DatasetRef, datasetRefName)
			err := e.Client.Status().Update(context.TODO(), newDataset)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

//Shutdown and clean up the engine
func (e *ReferenceDatasetEngine) Shutdown() (err error) {
	// 1. delete this dataset to mounted dataset DatasetRef field
	datasetRefName := getDatasetRefName(e.name, e.namespace)

	mountedRuntimeInfo, err := e.getMountedRuntimeInfo()
	if err != nil {
		return err
	}
	mountedDataset, err := utils.GetDataset(e.Client, mountedRuntimeInfo.GetName(), mountedRuntimeInfo.GetNamespace())
	if err != nil {
		return err
	}

	if utils.ContainsString(mountedDataset.Status.DatasetRef, datasetRefName) {
		newDataset := mountedDataset.DeepCopy()
		newDataset.Status.DatasetRef = utils.RemoveString(newDataset.Status.DatasetRef, datasetRefName)
		err := e.Client.Status().Update(context.TODO(), newDataset)
		if err != nil {
			return err
		}
	}
	return
}

func (e *ReferenceDatasetEngine) checkDatasetMountSupport() error {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	mountedNamespacedName := getMountedDatasetNamespacedName(dataset)
	// currently only support dataset mounting only one dataset
	if len(mountedNamespacedName) != 1 {
		return fmt.Errorf("ThinEngine with no profile name can only handle dataset only mounting one dataset")
	}

	mountedDataset, err := utils.GetDataset(e.Client, mountedNamespacedName[0].Name, mountedNamespacedName[0].Namespace)
	if err != nil {
		return err
	}
	// currently not support mounted dataset mounting another dataset
	if len(getMountedDatasetNamespacedName(mountedDataset)) != 0 {
		return fmt.Errorf("ThinEngine with no profile name can only handle dataset only mounting one dataset")
	}

	return nil
}
