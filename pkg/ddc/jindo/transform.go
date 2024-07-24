/*
Copyright 2022 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jindo

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	corev1 "k8s.io/api/core/v1"
)

func (e *JindoEngine) transform(runtime *datav1alpha1.JindoRuntime) (value *Jindo, err error) {
	if runtime == nil {
		err = fmt.Errorf("the jindoRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "JindoRuntime.Transform", "name", runtime.Name)

	if len(runtime.Spec.TieredStore.Levels) == 0 {
		err = fmt.Errorf("the TieredStore is null")
		return
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}

	var cachePaths []string // /mnt/disk1/bigboot or /mnt/disk1/bigboot,/mnt/disk2/bigboot
	stroagePath := runtime.Spec.TieredStore.Levels[0].Path
	originPath := strings.Split(stroagePath, ",")
	for _, value := range originPath {
		cachePaths = append(cachePaths, strings.TrimRight(value, "/")+"/"+
			e.namespace+"/"+e.name+"/bigboot")
	}
	metaPath := cachePaths[0]
	dataPath := strings.Join(cachePaths, ",")

	var userSetQuota []string // 1Gi or 1Gi,2Gi,3Gi
	if runtime.Spec.TieredStore.Levels[0].Quota != nil {
		userSetQuota = append(userSetQuota, utils.TransformQuantityToJindoUnit(runtime.Spec.TieredStore.Levels[0].Quota))
	}

	if runtime.Spec.TieredStore.Levels[0].QuotaList != "" {
		quotaList := runtime.Spec.TieredStore.Levels[0].QuotaList
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

	jindoSmartdataImage, smartdataTag, dnsServer := e.getSmartDataConfigs()
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
			HostPath: e.getHostMountPoint(),
		},
		Mounts: Mounts{
			Master:            e.transformMasterMountPath(metaPath),
			WorkersAndClients: e.transformWorkerMountPath(originPath),
		},
		Owner: transformer.GenerateOwnerReferenceFromObject(runtime),
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}
	e.transformNetworkMode(runtime, value)
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
	err = e.transformInitPortCheck(value)
	if err != nil {
		return
	}
	err = e.transformLabels(runtime, value)
	if err != nil {
		return
	}
	// set the placementMode
	e.transformPlacementMode(dataset, value)
	err = e.transformRunAsUser(runtime, value)
	if err != nil {
		return
	}
	e.transformTolerations(dataset, runtime, value)
	e.transformResources(runtime, value)
	e.transformLogConfig(runtime, value)
	value.Master.DnsServer = dnsServer
	value.Master.NameSpace = e.namespace
	value.Fuse.MountPath = jindoFuseMountpath
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
	for _, tmpMount := range dataset.Spec.Mounts {

		mount := tmpMount
		mount.Options = map[string]string{}
		mount.EncryptOptions = []datav1alpha1.EncryptOption{}

		for key, value := range dataset.Spec.SharedOptions {
			mount.Options[key] = value
		}
		for key, value := range tmpMount.Options {
			mount.Options[key] = value
		}

		mount.EncryptOptions = append(mount.EncryptOptions, dataset.Spec.SharedEncryptOptions...)
		mount.EncryptOptions = append(mount.EncryptOptions, tmpMount.EncryptOptions...)

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
			secret, err := kubeclient.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
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

	properties["storage.watermark.high.ratio"] = runtime.Spec.TieredStore.Levels[0].High
	//properties["storage.watermark.high.ratio"] = "0.4"

	properties["storage.watermark.low.ratio"] = runtime.Spec.TieredStore.Levels[0].Low
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

func (e *JindoEngine) transformResources(runtime *datav1alpha1.JindoRuntime, value *Jindo) {

	if runtime.Spec.Master.Resources.Limits != nil {
		e.Log.Info("setting Resources limit")
		if runtime.Spec.Master.Resources.Limits.Cpu() != nil {
			value.Master.Resources.Limits.CPU = runtime.Spec.Master.Resources.Limits.Cpu().String()
		}
		if runtime.Spec.Master.Resources.Limits.Memory() != nil {
			value.Master.Resources.Limits.Memory = runtime.Spec.Master.Resources.Limits.Memory().String()
		}
	}

	if runtime.Spec.Master.Resources.Requests != nil {
		e.Log.Info("setting Resources request")
		if runtime.Spec.Master.Resources.Requests.Cpu() != nil {
			value.Master.Resources.Requests.CPU = runtime.Spec.Master.Resources.Requests.Cpu().String()
		}
		if runtime.Spec.Master.Resources.Requests.Memory() != nil {
			value.Master.Resources.Requests.Memory = runtime.Spec.Master.Resources.Requests.Memory().String()
		}
	}

	if runtime.Spec.Fuse.Resources.Limits != nil {
		e.Log.Info("setting Resources limit")
		if runtime.Spec.Fuse.Resources.Limits.Cpu() != nil {
			value.Fuse.Resources.Limits.CPU = runtime.Spec.Fuse.Resources.Limits.Cpu().String()
		}
		if runtime.Spec.Fuse.Resources.Limits.Memory() != nil {
			value.Fuse.Resources.Limits.Memory = runtime.Spec.Fuse.Resources.Limits.Memory().String()
		}
	}

	if runtime.Spec.Fuse.Resources.Requests != nil {
		e.Log.Info("setting Resources request")
		if runtime.Spec.Fuse.Resources.Requests.Cpu() != nil {
			value.Fuse.Resources.Requests.CPU = runtime.Spec.Fuse.Resources.Requests.Cpu().String()
		}
		if runtime.Spec.Fuse.Resources.Requests.Memory() != nil {
			value.Fuse.Resources.Requests.Memory = runtime.Spec.Fuse.Resources.Requests.Memory().String()
		}
	}

	if runtime.Spec.Worker.Resources.Limits != nil {
		e.Log.Info("setting Resources limit")
		if runtime.Spec.Worker.Resources.Limits.Cpu() != nil {
			value.Worker.Resources.Limits.CPU = runtime.Spec.Worker.Resources.Limits.Cpu().String()
		}
		if runtime.Spec.Worker.Resources.Limits.Memory() != nil {
			value.Worker.Resources.Limits.Memory = runtime.Spec.Worker.Resources.Limits.Memory().String()
		}
	}

	if runtime.Spec.Worker.Resources.Requests != nil {
		e.Log.Info("setting Resources request")
		if runtime.Spec.Worker.Resources.Requests.Cpu() != nil {
			value.Worker.Resources.Requests.CPU = runtime.Spec.Worker.Resources.Requests.Cpu().String()
		}
		if runtime.Spec.Worker.Resources.Requests.Memory() != nil {
			value.Worker.Resources.Requests.Memory = runtime.Spec.Worker.Resources.Requests.Memory().String()
		}
	}
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
		"client.storage.connect.enable":             "true",
		"jfs.cache.meta-cache.enable":               "0",
		"jfs.cache.data-cache.enable":               "true",
		"jfs.cache.data-cache.slicecache.enable":    "true",
	}

	// "client.storage.rpc.port": "6101",
	properties["client.storage.rpc.port"] = strconv.Itoa(value.Worker.Port.Rpc)

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["jfs.cache.ram-cache.enable"] = "true"
	} else if e.getTieredStoreType(runtime) == 1 || e.getTieredStoreType(runtime) == 2 {
		// HDD and SSD
		properties["jfs.cache.ram-cache.enable"] = "false"
	}

	if len(runtime.Spec.Fuse.Properties) > 0 {
		for k, v := range runtime.Spec.Fuse.Properties {
			properties[k] = v
		}
	}
	value.Fuse.FuseProperties = properties

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()
	value.Fuse.HostPID = common.HostPIDEnabled(runtime.Annotations)

	return nil
}

func (e *JindoEngine) transformLogConfig(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	if len(runtime.Spec.LogConfig) > 0 {
		value.LogConfig = runtime.Spec.LogConfig
	} else {
		value.LogConfig = map[string]string{}
	}
}

func (e *JindoEngine) transformFuseNodeSelector(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}

	// The label will be added by CSI Plugin when any workload pod is scheduled on the node.
	value.Fuse.NodeSelector[e.getFuseLabelname()] = "true"

	return nil
}

func (e *JindoEngine) transformNodeSelector(runtime *datav1alpha1.JindoRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Worker.NodeSelector != nil {
		properties = runtime.Spec.Worker.NodeSelector
	}
	// } else {
	// 	labelName := e.getCommonLabelname()
	// 	properties[labelName] = "true"
	// }
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
	var baseArg = "-okernel_cache"
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

func (e *JindoEngine) getSmartDataConfigs() (image, tag, dnsServer string) {
	var (
		defaultImage     = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata"
		defaultTag       = "3.8.0"
		defaultDnsServer = "1.1.1.1"
	)

	image = docker.GetImageRepoFromEnv(common.JindoSmartDataImageEnv)
	tag = docker.GetImageTagFromEnv(common.JindoSmartDataImageEnv)
	dnsServer = os.Getenv(common.JindoDnsServer)
	if len(image) == 0 {
		image = defaultImage
	}
	if len(tag) == 0 {
		tag = defaultTag
	}
	if len(dnsServer) == 0 {
		dnsServer = defaultDnsServer
	}
	e.Log.Info("Set image", "image", image, "tag", tag, "dnsServer", dnsServer)

	return
}

func (e *JindoEngine) parseFuseImage() (image, tag string) {
	var (
		defaultImage = "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse"
		defaultTag   = "3.8.0"
	)

	image = docker.GetImageRepoFromEnv(common.JindoFuseImageEnv)
	tag = docker.GetImageTagFromEnv(common.JindoFuseImageEnv)
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

	// if not usehostnetwork then use default port
	// usehostnetwork to choose port from port allocator
	expectedPortNum := 2
	if !value.UseHostNetwork {
		value.Master.Port.Rpc = defaultMasterRpcRort
		value.Worker.Port.Rpc = DEFAULT_WORKER_RPC_PORT
		if value.Master.ReplicaCount == JINDO_HA_MASTERNUM {
			value.Master.Port.Raft = DEFAULT_RAFT_RPC_PORT
		}
		return nil
	}

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

func (e *JindoEngine) transformInitPortCheck(value *Jindo) error {
	// This function should be called after port allocation

	if !common.PortCheckEnabled() {
		return nil
	}

	e.Log.Info("Enabled port check")
	value.InitPortCheck.Enabled = true

	// Always use the default init image defined in env
	value.InitPortCheck.Image, value.InitPortCheck.ImageTag, value.InitPortCheck.ImagePullPolicy = docker.ParseInitImage("", "", "", common.DefaultInitImageEnv)

	// Inject ports to be checked to a init container which reports the usage status of the ports for easier debugging.
	// The jindo master container will always start even when some of the ports is in use.
	var ports []string

	ports = append(ports, strconv.Itoa(value.Master.Port.Rpc))
	if value.Master.ReplicaCount == JINDO_HA_MASTERNUM {
		ports = append(ports, strconv.Itoa(value.Master.Port.Raft))
	}

	// init container takes "PORT1:PORT2:PORT3..." as input
	value.InitPortCheck.PortsToCheck = strings.Join(ports, ":")

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
		value.Master.Tolerations = value.Tolerations
		value.Worker.Tolerations = value.Tolerations
		value.Fuse.Tolerations = value.Tolerations
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

func (e *JindoEngine) transformLabels(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	// the labels will not be merged here because they will be sequentially added into yaml templates
	// If two labels share the same label key, the last one in yaml templates overrides the former ones
	// and takes effect.
	value.Labels = runtime.Spec.Labels
	value.Master.Labels = runtime.Spec.Master.Labels
	value.Worker.Labels = runtime.Spec.Worker.Labels
	value.Fuse.Labels = runtime.Spec.Fuse.Labels

	return nil
}

func (e *JindoEngine) transformNetworkMode(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	// to set hostnetwork
	switch runtime.Spec.NetworkMode {
	case datav1alpha1.HostNetworkMode:
		value.UseHostNetwork = true
	case datav1alpha1.ContainerNetworkMode:
		value.UseHostNetwork = false
	case datav1alpha1.DefaultNetworkMode:
		value.UseHostNetwork = true
	}
}

func (e *JindoEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *Jindo) {

	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}
