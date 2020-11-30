package jindo

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"strings"
)

func (e *JindoEngine) transform(runtime *datav1alpha1.JindoRuntime) (value *Jindo, err error) {
	if runtime == nil {
		err = fmt.Errorf("The jindoRuntime is null")
		return
	}

	value = &Jindo{
		Image:           "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata",
		ImageTag:        "2.7.4",
		ImagePullPolicy: "Always",
		FuseImage:       "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse",
		FuseImageTag:    "2.7.4",
		User:            0,
		Group:           0,
		FsGroup:         0,
		UseHostNetwork:  true,
		UseHostPID:      true,
		Properties: map[string]string{
			"logDir": "/mnt/disk1/bigboot/log",
		},
		Master: Master{
			ReplicaCount: 1,
			Resources: Resources{
				Limits: Resource{
					CPU:    "4",
					Memory: "16G",
				},
				Requests: Resource{
					CPU:    "1",
					Memory: "1G",
				},
			},
			NodeSelector:     map[string]string{},
			MasterProperties: e.transformMaster(),
		},
		Worker: Worker{
			Resources: Resources{
				Limits: Resource{
					CPU:    "4",
					Memory: "8G",
				},
				Requests: Resource{
					CPU:    "1",
					Memory: "1G",
				},
			},
			NodeSelector:     map[string]string{},
			WorkerProperties: e.transformWorker(),
		},
		Fuse: Fuse{
			Args:           nil,
			HostPath:       "/mnt/jfs",
			NodeSelector:   map[string]string{},
			FuseProperties: e.transformFuse(),
		},
	}
	return value, nil

}

func (e *JindoEngine) transformMaster() map[string]string {
	properties := map[string]string{
		"namespace.rpc.port":            "8101",
		"namespace.meta-dir":            "/mnt/disk1/bigboot/server",
		"namespace.filelet.cache.size":  "100000",
		"namespace.blocklet.cache.size": "1000000",
		"namespace.backend.type":        "rocksdb",
	}
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return properties
	}
	jfsNamespace := ""
	for _, mount := range dataset.Spec.Mounts {
		if !strings.HasPrefix(mount.MountPoint, "oss://") {
			continue
		}
		jfsNamespace = jfsNamespace + mount.Name + ","
		properties["jfs.namespaces."+mount.Name+".oss.uri"] = mount.MountPoint
		properties["jfs.namespaces."+mount.Name+".oss.access.key"] = mount.Options["fs.oss.accessKeyId"]
		properties["jfs.namespaces."+mount.Name+".oss.access.secret"] = mount.Options["fs.oss.accessKeySecret"]
		properties["jfs.namespaces."+mount.Name+".mode"] = "cache"
	}
	if strings.HasSuffix(jfsNamespace, ",") {
		jfsNamespace = strings.TrimRight(jfsNamespace, ",")
	}
	properties["jfs.namespaces"] = jfsNamespace
	return properties
}

func (e *JindoEngine) transformWorker() map[string]string {
	properties := map[string]string{
		"storage.rpc.port":             "6101",
		"storage.data-dirs":            "/mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot",
		"storage.temp-data-dirs":       "/mnt/disk1/bigboot/tmp",
		"storage.watermark.high.ratio": "0.4",
		"storage.watermark.low.ratio":  "0.2",
		"storage.data-dirs.capacities": "80g,80g,80g",
		"storage.meta-dir":             "/mnt/disk1/bigboot/bignode",
	}
	return properties
}

func (e *JindoEngine) transformFuse() map[string]int {
	properties := map[string]int{
		"client.storage.rpc.port":                   6101,
		"client.oss.retry":                          5,
		"client.oss.upload.threads":                 4,
		"client.oss.upload.queue.size":              5,
		"client.oss.upload.max.parallelism":         16,
		"client.oss.timeout.millisecond":            30000,
		"client.oss.connection.timeout.millisecond": 3000,
	}
	return properties
}
