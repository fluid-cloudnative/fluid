package dataload

// DataLoadValue defines the value yaml file used in DataLoad helm chart
type DataLoadValue struct {
	DataLoadInfo DataLoadInfo `yaml:"dataloader"`
}

// DataLoadInfo defines values used in DataLoad helm chart
type DataLoadInfo struct {
	// BackoffLimit specifies the upper limit times when the DataLoad job fails
	BackoffLimit int32 `yaml:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `yaml:"targetDataset,omitempty"`

	// LoadMetadata specifies if the DataLoad job should load metadata from UFS when doing data load
	LoadMetadata bool `yaml:"loadMetadata,omitempty"`

	// TargetPaths specifies which paths should the DataLoad load
	TargetPaths []TargetPath `yaml:"targetPaths,omitempty"`

	// Image specifies the image that the DataLoad job uses
	Image string `yaml:"image,omitempty"`

	// CacheSmallData specifies if the dataload job should Cache Small Data
	CacheSmallData bool `yaml:"cacheSmallData,omitempty"`

	// CacheSmallData specifies if the dataload job should Cache Small Data
	LoadMemoryData bool `yaml:"loadMemoryData,omitempty"`

	// add HdfsConfig for JindoRuntime
	HdfsConfig string `yaml:"hdfsConfig,omitempty"`
}

type TargetPath struct {
	// Path specifies the path should be loaded
	Path string `yaml:"path,omitempty"`

	// Replicas specifies how many replicas should be loaded
	Replicas int32 `yaml:"replicas,omitempty"`

	// FluidNative specifies if the path is a native mountPoint(e.g. hostpath or pvc)
	FluidNative bool `yaml:"fluidNative,omitempty"`
}
