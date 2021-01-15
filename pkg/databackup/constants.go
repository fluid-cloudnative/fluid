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
	FINALIZER                = "fluid-databackup-controller-finalizer"
	BACPUP_PATH_POD          = "/alluxio_backups"
	DATABACKUP_CHART         = "fluid-databackup"
)
