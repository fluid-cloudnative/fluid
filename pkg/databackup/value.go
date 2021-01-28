package databackup

// DataBackupValue defines the value yaml file used in DataBackupValue helm chart
type DataBackupValue struct {
	DataBackupInfo DataBackupInfo `yaml:"databackuper"`
}

// DataBackupInfo defines values used in DataBackup helm chart
type DataBackupInfo struct {
	Namespace  string `yaml:"namespace,omitempty"`
	Dataset    string `yaml:"dataset,omitempty"`
	DataBackup string `yaml:"databackup,omitempty"`
	NodeName   string `yaml:"nodeName,omitempty"`
	Image      string `yaml:"image,omitempty"`
	JavaEnv    string `yaml:"javaEnv,omitempty"`
	Workdir    string `yaml:"workdir,omitempty"`
	PVCName    string `yaml:"pvcName,omitempty"`
	Path       string `yaml:"path,omitempty"`
}
