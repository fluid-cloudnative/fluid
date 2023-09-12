package base

import (
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetRemainingTimeToCleanUp return not nil remaining time if data operation has completion time and set ttlAfterFinished
func GetRemainingTimeToCleanUp(object client.Object, opStatus *datav1alpha1.OperationStatus, operateType datav1alpha1.OperationType) (*time.Duration, error) {
	if len(opStatus.Conditions) == 0 {
		// data operation has no completion time
		return nil, nil
	}

	ttl, err := GetTTL(object, operateType)
	if err != nil {
		return nil, err
	}
	if ttl == nil {
		// if data operation does not set ttlAfterFinished, return nil
		return nil, nil
	}

	curTime := time.Now()
	completionTime := opStatus.Conditions[0].LastProbeTime
	expireTime := completionTime.Add(time.Duration(*ttl) * time.Second)
	// calculate remainint time to clean up data operation
	remaining := expireTime.Sub(curTime)
	return &remaining, nil
}

func GetTTL(object client.Object, operateType datav1alpha1.OperationType) (ttl *int32, err error) {
	switch operateType {
	case datav1alpha1.DataBackupType:
		dataBackup, ok := object.(*datav1alpha1.DataBackup)
		if !ok {
			return ttl, fmt.Errorf("object %v is not a DataBackup", object)
		}
		ttl = dataBackup.Spec.TTLSecondsAfterFinished
	case datav1alpha1.DataLoadType:
		dataLoad, ok := object.(*datav1alpha1.DataLoad)
		if !ok {
			return ttl, fmt.Errorf("object %v is not a DataLoad", object)
		}
		if dataLoad.Spec.Policy == datav1alpha1.Cron {
			// do not clean up cron data operation
			return ttl, nil
		}
		ttl = dataLoad.Spec.TTLSecondsAfterFinished
	case datav1alpha1.DataMigrateType:
		dataMigrate, ok := object.(*datav1alpha1.DataMigrate)
		if !ok {
			return ttl, fmt.Errorf("object %v is not a DataMigrate", object)
		}
		if dataMigrate.Spec.Policy == datav1alpha1.Cron {
			return ttl, nil
		}
		ttl = dataMigrate.Spec.TTLSecondsAfterFinished
	case datav1alpha1.DataProcessType:
		dataProcess, ok := object.(*datav1alpha1.DataProcess)
		if !ok {
			return ttl, fmt.Errorf("object %v is not a DataProcess", object)
		}
		ttl = dataProcess.Spec.TTLSecondsAfterFinished
	}
	return ttl, nil
}
