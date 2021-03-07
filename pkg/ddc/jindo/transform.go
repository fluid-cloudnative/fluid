package jindo

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"regexp"
	"strings"
)

func (e *JindoEngine) transform(runtime *datav1alpha1.JindoRuntime) (value *Jindo, err error) {
	if runtime == nil {
		err = fmt.Errorf("the jindoRuntime is null")
		return
	}

	if len(runtime.Spec.Tieredstore.Levels) == 0 {
		err = fmt.Errorf("the Tieredstore is null")
		return
	}

	dataPath := runtime.Spec.Tieredstore.Levels[0].Path
	if strings.HasSuffix(dataPath, "/") {
		dataPath = strings.TrimRight(dataPath, "/")
	}

	value = &Jindo{
		Image:           "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata",
		ImageTag:        "3.3.5",
		ImagePullPolicy: "Always",
		FuseImage:       "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse",
		FuseImageTag:    "3.3.5",
		User:            0,
		Group:           0,
		FsGroup:         0,
		UseHostNetwork:  true,
		UseHostPID:      true,
		/*Properties:map[string]string{
			"logDir": "/mnt/disk1/bigboot/log",
		},*/
		Properties: e.transformPriority(dataPath),
		Master: Master{
			ReplicaCount:     1,
			NodeSelector:     map[string]string{},
			MasterProperties: e.transformMaster(runtime, dataPath),
		},
		Worker: Worker{
			NodeSelector:     e.transformNodeSelector(),
			WorkerProperties: e.transformWorker(runtime, dataPath),
		},
		Fuse: Fuse{
			Args:           e.transformFuseArg(),
			HostPath:       e.getMountPoint(),
			NodeSelector:   e.transformNodeSelector(),
			FuseProperties: e.transformFuse(runtime, dataPath),
		},
		Mounts: Mounts{
			Master:            e.transformMountPath(dataPath),
			WorkersAndClients: e.transformMountPath(dataPath),
		},
	}
	return value, nil
}

func (e *JindoEngine) transformMaster(runtime *datav1alpha1.JindoRuntime, dataPath string) map[string]string {
	properties := map[string]string{
		"namespace.rpc.port": "8101",
		//"namespace.meta-dir": "/mnt/disk1/bigboot/server",
		"namespace.filelet.cache.size":  "100000",
		"namespace.blocklet.cache.size": "1000000",
		"namespace.backend.type":        "rocksdb",
	}

	properties["namespace.meta-dir"] = dataPath + "/bigboot/server"

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
		properties["jfs.namespaces."+mount.Name+".oss.access.endpoint"] = mount.Options["fs.oss.endpoint"]

		if !strings.HasSuffix(mount.MountPoint, "/") {
			mount.MountPoint = mount.MountPoint + "/"
		}
		// transform mountpoint for jfs uri format
		var re = regexp.MustCompile(`(oss://(.*?))(/)`)
		rm := re.FindStringSubmatch(mount.MountPoint)
		if len(rm) < 2 {
			e.Log.Info("incorrect muountpath", "mount.MountPoint", mount.MountPoint)
		}
		mount.MountPoint = strings.Replace(mount.MountPoint, rm[1], rm[1]+"."+mount.Options["fs.oss.endpoint"], 1)

		properties["jfs.namespaces."+mount.Name+".oss.uri"] = mount.MountPoint
		properties["jfs.namespaces."+mount.Name+".mode"] = "cache"
		properties["jfs.namespaces."+mount.Name+".oss.access.key"] = mount.Options["fs.oss.accessKeyId"]
		properties["jfs.namespaces."+mount.Name+".oss.access.secret"] = mount.Options["fs.oss.accessKeySecret"]

		// to check whether encryptOptions exist
		for _, encryptOption := range mount.EncryptOptions {
			key := encryptOption.Name
			secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
			secret, err := utils.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
			if err != nil {
				e.Log.Info("can't get the secret")
				break
			}
			value := secret.Data[secretKeyRef.Key]
			if err != nil {
				e.Log.Info("decode value failed")
			}
			if key == "fs.oss.accessKeyId" {
				properties["jfs.namespaces."+mount.Name+".oss.access.key"] = string(value)
			}
			if key == "fs.oss.accessKeySecret" {
				properties["jfs.namespaces."+mount.Name+".oss.access.secret"] = string(value)
			}
			e.Log.Info("get from secret")
		}
	}
	if strings.HasSuffix(jfsNamespace, ",") {
		jfsNamespace = strings.TrimRight(jfsNamespace, ",")
	}
	properties["jfs.namespaces"] = jfsNamespace
	// combine properties together
	if len(runtime.Spec.Master.Properties) > 0 {
		for k, v := range runtime.Spec.Master.Properties {
			properties[k] = v
		}
	}
	return properties
}

func (e *JindoEngine) transformWorker(runtime *datav1alpha1.JindoRuntime, dataPath string) map[string]string {

	properties := map[string]string{
		"storage.rpc.port": "6101",
	}

	properties["namespace.meta-dir"] = dataPath + "/bigboot/bignode"

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["storage.ram.cache.size"] = utils.TranformQuantityToJindoUnit(runtime.Spec.Tieredstore.Levels[0].Quota)
		//properties["storage.ram.cache.size"] = "90g"

		properties["storage.slicelet.buffer.size"] = utils.TranformQuantityToJindoUnit(runtime.Spec.Tieredstore.Levels[0].Quota)
		//properties["storage.slicelet.buffer.size"] = "90g"
	}
	// HDD and SSD
	/*
	 spec:
	 replicas: 1
	 tieredstore:
	   levels:
	      - mediumtype: HDD
	       path: /mnt/disk1
	       quota: 240Gi
	       high: "0.4"
	       low: "0.2"
	*/
	properties["storage.data-dirs"] = dataPath + "/bigboot"
	//properties["storage.data-dirs"] = "/mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot"

	properties["storage.temp-data-dirs"] = dataPath + "/bigboot/tmp"
	//properties["storage.temp-data-dirs"] = "/mnt/disk1/bigboot/tmp"

	properties["storage.watermark.high.ratio"] = runtime.Spec.Tieredstore.Levels[0].High
	//properties["storage.watermark.high.ratio"] = "0.4"

	properties["storage.watermark.low.ratio"] = runtime.Spec.Tieredstore.Levels[0].Low
	//properties["storage.watermark.low.ratio"] = "0.2"

	properties["storage.data-dirs.capacities"] = utils.TranformQuantityToJindoUnit(runtime.Spec.Tieredstore.Levels[0].Quota)
	///properties["storage.data-dirs.capacities"] = "80g,80g,80g"

	if len(runtime.Spec.Worker.Properties) > 0 {
		for k, v := range runtime.Spec.Worker.Properties {
			properties[k] = v
		}
	}
	return properties
}

func (e *JindoEngine) transformFuse(runtime *datav1alpha1.JindoRuntime, dataPath string) map[string]string {
	// default enable data-cache and disable meta-cache
	properties := map[string]string{
		"client.storage.rpc.port":                   "6101",
		"client.oss.retry":                          "5",
		"client.oss.upload.threads":                 "4",
		"client.oss.upload.queue.size":              "5",
		"client.oss.upload.max.parallelism":         "16",
		"client.oss.timeout.millisecond":            "30000",
		"client.oss.connection.timeout.millisecond": "3000",
		"jfs.cache.meta-cache.enable":               "0",
		"jfs.cache.data-cache.enable":               "1",
		"jfs.cache.data-cache.slicecache.enable":    "1",
	}

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["jfs.cache.ram-cache.enable"] = "1"
	} else if e.getTieredStoreType(runtime) == 1 || e.getTieredStoreType(runtime) == 2 {
		// HDD and SSD
		properties["jfs.cache.ram-cache.enable"] = "0"
	}

	if len(runtime.Spec.Fuse.Properties) > 0 {
		for k, v := range runtime.Spec.Fuse.Properties {
			properties[k] = v
		}
	}
	return properties
}

func (e *JindoEngine) transformNodeSelector() map[string]string {
	labelName := e.getCommonLabelname()
	properties := map[string]string{}
	properties[labelName] = "true"
	return properties
}

func (e *JindoEngine) transformPriority(dataPath string) map[string]string {
	properties := map[string]string{}
	properties["logDir"] = dataPath + "/bigboot/log"
	return properties
}

func (e *JindoEngine) transformMountPath(dataPath string) map[string]string {
	properties := map[string]string{}
	properties["1"] = dataPath
	return properties
}

func (e *JindoEngine) transformFuseArg() []string {
	if len(e.runtime.Spec.Fuse.Args) > 0 {
		return e.runtime.Spec.Fuse.Args
	}
	return []string{"-okernel_cache -oattr_timeout=9000 -oentry_timeout=9000"}
}
