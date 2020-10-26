package dataload

//todo(xuzhihao): Add comments on this file.

// DataLoadValue defines the value yaml file used in DataLoad helm chart
type DataLoadValue struct {
	DataLoadInfo DataLoadInfo `yaml:"dataloader"`
}

type DataLoadInfo struct {
	BackOffLimit int32 `yaml:"backoffLimit,omitempty"`

	TargetDataset string `yaml:"targetDataset,omitempty"`

	LoadMetadata bool `yaml:"loadMetadata,omitempty"`

	TargetPaths []TargetPath `yaml:"targetPaths,omitempty"`

	Image string `yaml:"image,omitempty"`
}

type TargetPath struct {
	Path        string `yaml:"path,omitempty"`
	Replicas    int32  `yaml:"replicas,omitempty"`
	FluidNative bool   `yaml:"fluidNative,omitempty"`
}
