package databackup

import "github.com/fluid-cloudnative/fluid/pkg/utils/docker"

// DataBackupValue defines the value yaml file used in DataBackup helm chart
type DataBackupValue struct {
	DataBackup      DataBackup `yaml:"dataBackup"`
	docker.UserInfo `yaml:",inline"`
	InitUsers       docker.InitUsers `yaml:"initUsers,omitempty"`
}

// DataBackup defines values used in DataBackup helm chart
type DataBackup struct {
	Namespace string `yaml:"namespace,omitempty"`
	Dataset   string `yaml:"dataset,omitempty"`
	Name      string `yaml:"name,omitempty"`
	NodeName  string `yaml:"nodeName,omitempty"`
	Image     string `yaml:"image,omitempty"`
	JavaEnv   string `yaml:"javaEnv,omitempty"`
	Workdir   string `yaml:"workdir,omitempty"`
	PVCName   string `yaml:"pvcName,omitempty"`
	Path      string `yaml:"path,omitempty"`
}
