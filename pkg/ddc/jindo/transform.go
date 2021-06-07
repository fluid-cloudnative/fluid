package jindo

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strconv"
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

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}

	var cachePaths []string // /mnt/disk1/bigboot or /mnt/disk1/bigboot,/mnt/disk2/bigboot
	stroagePath := runtime.Spec.Tieredstore.Levels[0].Path
	originPath := strings.Split(stroagePath, ",")
	for _, value := range originPath {
		cachePaths = append(cachePaths, strings.TrimRight(value, "/")+"/"+
			e.namespace+"/"+e.name+"/bigboot")
	}
	metaPath := cachePaths[0]
	dataPath := strings.Join(cachePaths, ",")

	var userSetQuota []string // 1Gi or 1Gi,2Gi,3Gi
	if runtime.Spec.Tieredstore.Levels[0].Quota != nil {
		userSetQuota = append(userSetQuota, utils.TranformQuantityToJindoUnit(runtime.Spec.Tieredstore.Levels[0].Quota))
	}

	if runtime.Spec.Tieredstore.Levels[0].QuotaList != "" {
		quotaList := runtime.Spec.Tieredstore.Levels[0].QuotaList
		quotas := strings.Split(quotaList, ",")
		if len(quotas) != len(originPath) {
			err = fmt.Errorf("the num of cache path and quota must be equal")
			return
		}
		for _, value := range quotas {
			if strings.HasSuffix(value, "Gi") {
				value = strings.ReplaceAll(value, "Gi", "g")
			}
			userSetQuota = append(userSetQuota, value)
		}
	}
	userQuotas := strings.Join(userSetQuota, ",") // 1g or 1g,2g

	jindoSmartdataImage, smartdataTag := e.parseSmartDataImage()
	jindoFuseImage, fuseTag := e.parseFuseImage()

	value = &Jindo{
		Image:           jindoSmartdataImage,
		ImageTag:        smartdataTag,
		ImagePullPolicy: "Always",
		FuseImage:       jindoFuseImage,
		FuseImageTag:    fuseTag,
		User:            0,
		Group:           0,
		FsGroup:         0,
		UseHostNetwork:  true,
		UseHostPID:      true,
		Properties:      e.transformPriority(metaPath),
		Master: Master{
			ReplicaCount: e.transformReplicasCount(runtime),
			NodeSelector: e.transformMasterSelector(runtime),
		},
		Worker: Worker{
			NodeSelector: e.transformNodeSelector(runtime),
		},
		Fuse: Fuse{
			Args:     e.transformFuseArg(runtime, dataset),
			HostPath: e.getMountPoint(),
		},
		Mounts: Mounts{
			Master:            e.transformMasterMountPath(metaPath),
			WorkersAndClients: e.transformWorkerMountPath(originPath),
		},
	}
	err = e.transformHadoopConfig(runtime, value)
	if err != nil {
		return
	}
	err = e.transformFuseNodeSelector(runtime, value)
	if err != nil {
		return
	}
	err = e.transformSecret(runtime, value)
	if err != nil {
		return
	}
	err = e.transformToken(runtime, value)
	if err != nil {
		return
	}
	err = e.allocatePorts(value)
	if err != nil {
		return
	}
	err = e.transformMaster(runtime, metaPath, value, dataset)
	if err != nil {
		return
	}
	err = e.transformWorker(runtime, metaPath, dataPath, userQuotas, value)
	if err != nil {
		return
	}
	err = e.transformFuse(runtime, value)
	if err != nil {
		return
	}
	err = e.transformRunAsUser(runtime, value)
	e.transformTolerations(dataset, runtime, value)
	return value, err
}

func (e *JindoEngine) transformMaster(runtime *datav1alpha1.JindoRuntime, metaPath string, value *Jindo, dataset *datav1alpha1.Dataset) (err error) {
	properties := map[string]string{
		//"namespace.meta-dir": "/mnt/disk1/bigboot/server",
		"namespace.filelet.cache.size":  "100000",
		"namespace.blocklet.cache.size": "1000000",
		"namespace.backend.type":        "rocksdb",
	}

	if value.Master.ReplicaCount == 3 {
		properties["namespace.backend.type"] = "raft"
	}

	//"namespace.rpc.port": "8101",
	properties["namespace.rpc.port"] = strconv.Itoa(value.Master.Port.Rpc)

	properties["namespace.meta-dir"] = metaPath + "/server"

	jfsNamespace := "jindo"
	mode := "oss"
	for _, mount := range dataset.Spec.Mounts {

		//jfsNamespace = jfsNamespace + mount.Name + ","

		if !strings.HasSuffix(mount.MountPoint, "/") {
			mount.MountPoint = mount.MountPoint + "/"
		}
		// transform mountpoint for oss or hdfs format
		if strings.HasPrefix(mount.MountPoint, "hdfs://") {
			properties["jfs.namespaces.jindo.hdfs.uri"] = mount.MountPoint
			mode = "hdfs"
		} else if strings.HasPrefix(mount.MountPoint, "s3://") {
			properties["jfs.namespaces.jindo.s3.uri"] = mount.MountPoint
			properties["jfs.namespaces.jindo.s3.access.key"] = mount.Options["fs.s3.accessKeyId"]
			properties["jfs.namespaces.jindo.s3.access.secret"] = mount.Options["fs.s3.accessKeySecret"]
			mode = "s3"
		} else {
			if !strings.HasPrefix(mount.MountPoint, "oss://") {
				continue
			}

			var re = regexp.MustCompile(`(oss://(.*?))(/)`)
			rm := re.FindStringSubmatch(mount.MountPoint)
			if len(rm) < 2 {
				e.Log.Info("incorrect muountpath", "mount.MountPoint", mount.MountPoint)
			}
			mount.MountPoint = strings.Replace(mount.MountPoint, rm[1], rm[1]+"."+mount.Options["fs.oss.endpoint"], 1)
			properties["jfs.namespaces.jindo.oss.uri"] = mount.MountPoint
			properties["jfs.namespaces.jindo.oss.access.key"] = mount.Options["fs.oss.accessKeyId"]
			properties["jfs.namespaces.jindo.oss.access.secret"] = mount.Options["fs.oss.accessKeySecret"]
			properties["jfs.namespaces.jindo.oss.access.endpoint"] = mount.Options["fs.oss.endpoint"]
		}
		properties["jfs.namespaces.jindo.mode"] = "cache"
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
			if key == "fs."+mode+".accessKeyId" {
				properties["jfs.namespaces.jindo."+mode+".access.key"] = string(value)
			}
			if key == "fs."+mode+".accessKeySecret" {
				properties["jfs.namespaces.jindo."+mode+".access.secret"] = string(value)
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

	if mode == "oss" || mode == "s3" {
		value.Master.OssKey = properties["jfs.namespaces.jindo."+mode+".access.key"]
		value.Master.OssSecret = properties["jfs.namespaces.jindo."+mode+".access.secret"]
	}
	value.Master.MasterProperties = properties

	return nil
}

func (e *JindoEngine) transformWorker(runtime *datav1alpha1.JindoRuntime, metaPath string, dataPath string, userQuotas string, value *Jindo) (err error) {

	properties := map[string]string{}
	// "storage.rpc.port": "6101",
	properties["storage.rpc.port"] = strconv.Itoa(value.Worker.Port.Rpc)

	properties["namespace.meta-dir"] = metaPath + "/bignode"

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["storage.ram.cache.size"] = userQuotas
		//properties["storage.ram.cache.size"] = "90g"

		properties["storage.slicelet.buffer.size"] = userQuotas
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
	properties["storage.data-dirs"] = dataPath
	//properties["storage.data-dirs"] = "/mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot"

	properties["storage.temp-data-dirs"] = metaPath + "/tmp"
	//properties["storage.temp-data-dirs"] = "/mnt/disk1/bigboot/tmp"

	properties["storage.watermark.high.ratio"] = runtime.Spec.Tieredstore.Levels[0].High
	//properties["storage.watermark.high.ratio"] = "0.4"

	properties["storage.watermark.low.ratio"] = runtime.Spec.Tieredstore.Levels[0].Low
	//properties["storage.watermark.low.ratio"] = "0.2"

	properties["storage.data-dirs.capacities"] = userQuotas
	///properties["storage.data-dirs.capacities"] = "80g,80g,80g"

	if len(runtime.Spec.Worker.Properties) > 0 {
		for k, v := range runtime.Spec.Worker.Properties {
			properties[k] = v
		}
	}
	value.Worker.WorkerProperties = properties
	return nil
}

func (e *JindoEngine) transformFuse(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	// default enable data-cache and disable meta-cache
	properties := map[string]string{
		"client.oss.retry":                          "5",
		"client.oss.upload.threads":                 "4",
		"client.oss.upload.queue.size":              "5",
		"client.oss.upload.max.parallelism":         "16",
		"client.oss.timeout.millisecond":            "30000",
		"client.oss.connection.timeout.millisecond": "3000",
		"jfs.cache.meta-cache.enable":               "0",
		"jfs.cache.data-cache.enable":               "1",
		"jfs.cache.data-cache.slicecache.enable":    "0",
	}

	// "client.storage.rpc.port": "6101",
	properties["client.storage.rpc.port"] = strconv.Itoa(value.Worker.Port.Rpc)

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
	value.Fuse.FuseProperties = properties
	return nil
}

func (e *JindoEngine) transformFuseNodeSelector(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	value.Fuse.NodeSelector = map[string]string{}
	if runtime.Spec.Fuse.Global {
		value.Fuse.Global = true
		if len(runtime.Spec.Fuse.NodeSelector) > 0 {
			value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
		}
		value.Fuse.NodeSelector[common.FLUID_FUSE_BALLOON_KEY] = common.FLUID_FUSE_BALLOON_VALUE
		e.Log.Info("Enable Fuse's global mode")
	} else {
		if runtime.Spec.Fuse.NodeSelector == nil {
			labelName := e.getCommonLabelname()
			value.Fuse.NodeSelector[labelName] = "true"
			e.Log.Info("Disable Fuse's global mode")
		} else {
			value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
		}
	}
	return nil
}

func (e *JindoEngine) transformNodeSelector(runtime *datav1alpha1.JindoRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Worker.NodeSelector != nil {
		properties = runtime.Spec.Worker.NodeSelector
	} else {
		labelName := e.getCommonLabelname()
		properties[labelName] = "true"
	}
	return properties
}

func (e *JindoEngine) transformReplicasCount(runtime *datav1alpha1.JindoRuntime) int {
	if runtime.Spec.Master.Replicas == JINDO_HA_MASTERNUM {
		return JINDO_HA_MASTERNUM
	}
	return JINDO_MASTERNUM_DEFAULT
}

func (e *JindoEngine) transformMasterSelector(runtime *datav1alpha1.JindoRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Master.NodeSelector != nil {
		properties = runtime.Spec.Master.NodeSelector
	}
	return properties
}

func (e *JindoEngine) transformPriority(metaPath string) map[string]string {
	properties := map[string]string{}
	properties["logDir"] = metaPath + "/log"
	return properties
}

func (e *JindoEngine) transformMasterMountPath(metaPath string) map[string]string {
	properties := map[string]string{}
	properties["1"] = metaPath
	return properties
}

func (e *JindoEngine) transformWorkerMountPath(originPath []string) map[string]string {
	properties := map[string]string{}
	for index, value := range originPath {
		properties[strconv.Itoa(index+1)] = strings.TrimRight(value, "/")
	}
	return properties
}

func (e *JindoEngine) transformFuseArg(runtime *datav1alpha1.JindoRuntime, dataset *datav1alpha1.Dataset) []string {
	var baseArg = "-okernel_cache -oattr_timeout=9000 -oentry_timeout=9000"
	var rootArg = ""
	var secretArg = ""
	if len(dataset.Spec.Mounts) > 0 && dataset.Spec.Mounts[0].Path != "" {
		rootArg = "-oroot_ns=jindo"
		baseArg = rootArg + " " + baseArg
	}
	if len(runtime.Spec.Secret) != 0 {
		secretArg = "-ocredential_provider=secrets:///token/"
		baseArg = secretArg + " " + baseArg
	}

	if len(runtime.Spec.Fuse.Args) > 0 {
		properties := runtime.Spec.Fuse.Args
		if rootArg != "" {
			properties = append(properties, rootArg)
		}
		if len(runtime.Spec.Secret) != 0 {
			properties = append(properties, secretArg)
		}
		return properties
	}
	return []string{baseArg}
}

func (e *JindoEngine) parseSmartDataImage() (image, tag string) {
	var (
		defaultImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata"
		defaultTag   = "3.5.1"
	)

	image = docker.GetImageRepoFromEnv(common.JINDO_SMARTDATA_IMAGE_ENV)
	tag = docker.GetImageTagFromEnv(common.JINDO_SMARTDATA_IMAGE_ENV)
	if len(image) == 0 {
		image = defaultImage
	}
	if len(tag) == 0 {
		tag = defaultTag
	}
	e.Log.Info("Set image", "image", image, "tag", tag)

	return
}

func (e *JindoEngine) parseFuseImage() (image, tag string) {
	var (
		defaultImage = "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse"
		defaultTag   = "3.5.1"
	)

	image = docker.GetImageRepoFromEnv(common.JINDO_FUSE_IMAGE_ENV)
	tag = docker.GetImageTagFromEnv(common.JINDO_FUSE_IMAGE_ENV)
	if len(image) == 0 {
		image = defaultImage
	}
	if len(tag) == 0 {
		tag = defaultTag
	}
	e.Log.Info("Set image", "image", image, "tag", tag)

	return
}

func (e *JindoEngine) transformSecret(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	if len(runtime.Spec.Secret) != 0 {
		value.Secret = runtime.Spec.Secret
	}
	return nil
}

func (e *JindoEngine) transformToken(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	if len(runtime.Spec.Secret) != 0 {
		properties := map[string]string{
			"default.credential.provider": "secrets:///token/",
		}
		value.Master.TokenProperties = properties
	}
	return nil
}

func (e *JindoEngine) allocatePorts(value *Jindo) error {
	// For now, Jindo only needs two ports for master and client rpc respectively
	expectedPortNum := 2

	if value.Master.ReplicaCount == JINDO_HA_MASTERNUM {
		expectedPortNum = 3
	}

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		e.Log.Error(err, "can't get runtime port allocator")
		return err
	}

	allocatedPorts, err := allocator.GetAvailablePorts(expectedPortNum)
	if err != nil {
		e.Log.Error(err, "can't get available ports", "expected port num", expectedPortNum)
		return err
	}

	index := 0
	value.Master.Port.Rpc = allocatedPorts[index]
	index++
	value.Worker.Port.Rpc = allocatedPorts[index]
	if value.Master.ReplicaCount == JINDO_HA_MASTERNUM {
		index++
		value.Master.Port.Raft = allocatedPorts[index]
	}
	return nil
}

func (e *JindoEngine) transformRunAsUser(runtime *datav1alpha1.JindoRuntime, value *Jindo) error {
	if len(runtime.Spec.User) != 0 {
		value.Fuse.RunAs = runtime.Spec.User
	}
	return nil
}

func (e *JindoEngine) transformTolerations(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.JindoRuntime, value *Jindo) {

	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}

	if len(runtime.Spec.Master.Tolerations) > 0 {
		for _, toleration := range runtime.Spec.Master.Tolerations {
			toleration.TolerationSeconds = nil
			value.Master.Tolerations = append(value.Tolerations, toleration)
		}
	}

	if len(runtime.Spec.Worker.Tolerations) > 0 {
		for _, toleration := range runtime.Spec.Worker.Tolerations {
			toleration.TolerationSeconds = nil
			value.Worker.Tolerations = append(value.Tolerations, toleration)
		}
	}

	if len(runtime.Spec.Fuse.Tolerations) > 0 {
		for _, toleration := range runtime.Spec.Fuse.Tolerations {
			toleration.TolerationSeconds = nil
			value.Fuse.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
