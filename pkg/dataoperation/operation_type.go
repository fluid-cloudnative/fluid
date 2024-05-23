package dataoperation

type OperationType string

const (
	DataLoadType    OperationType = "DataLoad"
	DataBackupType  OperationType = "DataBackup"
	DataMigrateType OperationType = "DataMigrate"
	DataProcessType OperationType = "DataProcess"
)

func (o OperationType) getPods() {

}
