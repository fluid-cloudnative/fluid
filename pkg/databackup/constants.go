package databackup

type Phase string

const (
	PhaseNone      Phase = ""
	PhasePending   Phase = "Pending"
	PhaseBackuping Phase = "Backuping"
	PhaseComplete  Phase = "Complete"
	PhaseFailed    Phase = "Failed"
)

// ConditionType is a valid value for DataBackupCondition.Type
type ConditionType string

// These are valid conditions of a DataBackup.
const (
	// Complete means the DataBackup has completed its execution.
	Complete ConditionType = "Complete"
	// Failed means the DataBackup has failed its execution.
	Failed ConditionType = "Failed"
)

const (
	FINALIZER          = "fluid-databackup-controller-finalizer"
	DATABACKUP_IMA_URL = "registry.cn-hangzhou.aliyuncs.com/fluid/fluid-databackup"
	DATABACKUP_IMA_TAG = ":v0.5.0-c07460a"
	BACKUP_RESULT_BACKUP_URI = "Backup URI         : "
	BACPUP_PATH_HOST = "/alluxio_backups"
	BACPUP_PATH_POD = "/alluxio_backups"
	PVC_PATH_POD = "/pvc"
	BACKUP_CONTAINER_NAME = "tool"


)
