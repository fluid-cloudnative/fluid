package alluxio

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
)

// MetadataSyncResult describes result for asynchronous metadata sync
type MetadataSyncResult struct {
	Done      bool
	StartTime time.Time
	UfsTotal  string
	FileNum   string
	Err       error
}

// SyncMetadata syncs metadata if necessary
// For Alluxio Engine, metadata sync is an asynchronous operation, which means
// you should call this function periodically to make sure the function actually takes effect.
func (e *AlluxioEngine) SyncMetadata() (err error) {
	should, err := e.shouldSyncMetadata()
	if err != nil {
		e.Log.Error(err, "Failed to check if should sync metadata")
		return
	}
	if should {
		return e.syncMetadataInternal()
	}
	return
}

// shouldSyncMetadata checks dataset's UfsTotal to decide whether should sync metadata
func (e *AlluxioEngine) shouldSyncMetadata() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		should = false
		return should, err
	}

	//todo(xuzhihao): option to enable/disable automatic metadata sync
	//todo: periodical metadata sync
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
		e.Log.V(1).Info("dataset ufs is ready",
			"dataset name", dataset.Name,
			"dataset namespace", dataset.Namespace,
			"ufstotal", dataset.Status.UfsTotal)
		should = false
		return should, nil
	}
	should = true
	return should, nil
}

// syncMetadataInternal do the actual work of metadata sync
// At any time, there is at most one goroutine working on metadata sync. First call to
// this function will start a goroutine including the following two steps:
//   1. load metadata
//   2. get total size of UFSs
// Any following calls to this function will try to get result of the working goroutine with a timeout, which
// ensures the function won't block the following Sync operations(e.g. CheckAndUpdateRuntimeStatus) for a long time.
func (e *AlluxioEngine) syncMetadataInternal() (err error) {
	// init a MetadataInfoFile been saved in alluxio-master
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	metadataInfoFile := e.GetMetadataInfoFile()
	err = fileUtils.InitMetadataInfoFile(e.name, metadataInfoFile)
	if err != nil {
		e.Log.Error(err, "Failed to InitMetadataInfoFile of the dataset")
	}

	if e.MetadataSyncDoneCh != nil {
		// Either get result from channel or timeout
		select {
		case result := <-e.MetadataSyncDoneCh:
			defer func() {
				e.MetadataSyncDoneCh = nil
			}()
			e.Log.Info("Get result from MetadataSyncDoneCh", "result", result)
			if result.Done {
				e.Log.Info("Metadata sync succeeded", "period", time.Since(result.StartTime))
				err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
					dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
					if err != nil {
						return
					}
					datasetToUpdate := dataset.DeepCopy()
					datasetToUpdate.Status.UfsTotal = result.UfsTotal
					datasetToUpdate.Status.FileNum = result.FileNum
					if !reflect.DeepEqual(datasetToUpdate, dataset) {
						err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
						if err != nil {
							return
						}
					}
					// backup the ufs total result in local
					err = fileUtils.InsertMetaDataInfoIntoFile(operations.UfsTotal, result.UfsTotal, metadataInfoFile)
					if err != nil {
						e.Log.Error(err, "Failed to InsertMetaDataInfoIntoFile of the dataset")
					}
					// backup the file num result in local
					err = fileUtils.InsertMetaDataInfoIntoFile(operations.FileNum, result.FileNum, metadataInfoFile)
					if err != nil {
						e.Log.Error(err, "Failed to InsertMetaDataInfoIntoFile of the dataset")
					}
					return
				})
				if err != nil {
					e.Log.Error(err, "Failed to update UfsTotal and FileNum of the dataset")
					return err
				}
			} else {
				e.Log.Error(result.Err, "Metadata sync failed")
				return result.Err
			}
		case <-time.After(CHECK_METADATA_SYNC_DONE_TIMEOUT_MILLISEC * time.Millisecond):
			e.Log.V(1).Info("Metadata sync still in progress")
		}
	} else {
		// Metadata sync haven't started
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
			if err != nil {
				return
			}
			datasetToUpdate := dataset.DeepCopy()
			datasetToUpdate.Status.UfsTotal = METADATA_SYNC_NOT_DONE_MSG
			datasetToUpdate.Status.FileNum = METADATA_SYNC_NOT_DONE_MSG
			if !reflect.DeepEqual(dataset, datasetToUpdate) {
				err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
				if err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			e.Log.Error(err, "Failed to set UfsTotal to METADATA_SYNC_NOT_DONE_MSG")
		}
		e.MetadataSyncDoneCh = make(chan MetadataSyncResult)
		go func(resultChan chan MetadataSyncResult) {
			defer close(resultChan)
			result := MetadataSyncResult{
				StartTime: time.Now(),
				UfsTotal:  "",
			}
			dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
			if err != nil {
				e.Log.Error(err, "Can't get dataset when syncing metadata", "name", e.name, "namespace", e.namespace)
				result.Err = err
				result.Done = false
				resultChan <- result
				return
			}

			e.Log.Info("Metadata Sync starts", "dataset namespace", e.namespace, "dataset name", e.name)

			metadataInfoRestoreFile := ""
			if dataset.Spec.DataRestoreLocation != nil {
				if dataset.Spec.DataRestoreLocation.Path != "" {
					pvcName, path, err := utils.ParseBackupRestorePath(dataset.Spec.DataRestoreLocation.Path)
					if err != nil {
						e.Log.Error(err, "restore path cannot analyse", "Path", dataset.Spec.DataRestoreLocation.Path)
					} else {
						if pvcName != "" {
							metadataInfoRestoreFile = "/pvc" + path + e.GetMetadataInfoFileName()
						} else {
							metadataInfoRestoreFile = "/host/" + e.GetMetadataInfoFileName()
						}
					}
				}
			}

			podName, containerName := e.getMasterPodInfo()
			fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

			if metadataInfoRestoreFile == "" {
				// sync local dir if necessary
				for _, mount := range dataset.Spec.Mounts {
					if e.isFluidNativeScheme(mount.MountPoint) {
						localDirPath := fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), mount.Name)
						e.Log.Info(fmt.Sprintf("Syncing local dir, path: %s", localDirPath))
						err = fileUtils.SyncLocalDir(localDirPath)
						if err != nil {
							e.Log.Error(err, fmt.Sprintf("Sync local dir failed when syncing metadata, path: %s", localDirPath), "name", e.name, "namespace", e.namespace)
							result.Err = err
							result.Done = false
							resultChan <- result
							return
						}
					}
				}

				// load metadata
				err = fileUtils.LoadMetadataWithoutTimeout("/")
				if err != nil {
					e.Log.Error(err, "LoadMetadata failed when syncing metadata", "name", e.name, "namespace", e.namespace)
					result.Err = err
					result.Done = false
					resultChan <- result
					return
				}
			}
			result.Done = true

			shouldGetUfsTotal := true
			shouldGetFileNum := true
			// restore metadata info from restore file
			if metadataInfoRestoreFile != "" {
				result.UfsTotal, err = fileUtils.QueryMetaDataInfoIntoFile(operations.UfsTotal, metadataInfoRestoreFile)
				if err == nil && result.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
					shouldGetUfsTotal = false
				}
				result.FileNum, err = fileUtils.QueryMetaDataInfoIntoFile(operations.FileNum, metadataInfoRestoreFile)
				if err == nil && result.FileNum != METADATA_SYNC_NOT_DONE_MSG {
					shouldGetFileNum = false
				}
			}
			// get Ufs Total size
			if shouldGetUfsTotal {
				datasetUFSTotalBytes, err := e.TotalStorageBytes()
				if err != nil {
					e.Log.Error(err, "Get Ufs Total size failed when syncing metadata", "name", e.name, "namespace", e.namespace)
					result.Done = false
				} else {
					result.UfsTotal = utils.BytesSize(float64(datasetUFSTotalBytes))
				}
			}
			// get File Num
			if shouldGetFileNum {
				fileNum, err := e.getDataSetFileNum()
				if err != nil {
					e.Log.Error(err, "Get File Num failed when syncing metadata", "name", e.name, "namespace", e.namespace)
					result.Done = false
				} else {
					result.FileNum = fileNum
				}
			}

			if !result.Done {
				result.Err = errors.New("GetMetadataInfoFailed")
			} else {
				result.Err = nil
			}
			resultChan <- result
		}(e.MetadataSyncDoneCh)
	}
	return
}
