package jindo

type Jindo struct {
	Image           string            `yaml:"image"`
	ImageTag        string            `yaml:"imageTag"`
	ImagePullPolicy string            `yaml:"imagePullPolicy"`
	FuseImage       string            `yaml:"fuseImage"`
	FuseImageTag    string            `yaml:"fuseImageTag"`
	User            int               `yaml:"user"`
	Group           int               `yaml:"group"`
	FsGroup         int               `yaml:"fsGroup"`
	UseHostNetwork  bool              `yaml:"useHostNetwork"`
	UseHostPID      bool              `yaml:"useHostPID"`
	Properties      map[string]string `yaml:"properties"`
	Master          Master            `yaml:"master"`
	Worker          Worker            `yaml:"worker"`
	Fuse            Fuse              `yaml:"fuse"`
	Mounts          Mounts            `yaml:"mounts"`
}

type Master struct {
	ReplicaCount     int               `yaml:"replicaCount"`
	Resources        Resources         `yaml:"resources"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	MasterProperties map[string]string `yaml:"properties"`
}

type Worker struct {
	Resources        Resources         `yaml:"resources"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	WorkerProperties map[string]string `yaml:"properties"`
}

type Fuse struct {
	Args           []string          `yaml:"args"`
	HostPath       string            `yaml:"hostPath"`
	NodeSelector   map[string]string `yaml:"nodeSelector,omitempty"`
	FuseProperties map[string]string `yaml:"properties"`
}

type Mounts struct {
	Master            map[string]string `yaml:"master"`
	WorkersAndClients map[string]string `yaml:"workersAndClients"`
}

type Resources struct {
	Limits   Resource `yaml:"limits"`
	Requests Resource `yaml:"requests"`
}

type Resource struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type cacheStates struct {
	cacheCapacity string
	// cacheable        string
	// lowWaterMark     string
	// highWaterMark    string
	cached           string
	cachedPercentage string
	// nonCacheable     string
}
