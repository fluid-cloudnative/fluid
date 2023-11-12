/*
Copyright 2023 The Fluid Authors.

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

package jindofsx

import (
	"context"
	"errors"
	"fmt"
	versionutil "github.com/fluid-cloudnative/fluid/pkg/utils/version"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
)

type smartdataConfig struct {
	image           string
	imageTag        string
	imagePullPolicy string
	dnsServer       string
}

func (e *JindoFSxEngine) transform(runtime *datav1alpha1.JindoRuntime) (value *Jindo, err error) {
	if runtime == nil {
		err = fmt.Errorf("the jindoRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "JindoRuntime.Transform", "name", runtime.Name)

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}

	var cachePaths []string // /mnt/disk1/bigboot or /mnt/disk1/bigboot,/mnt/disk2/bigboot
	var stroagePath = "/dev/shm/"
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		stroagePath = runtime.Spec.TieredStore.Levels[0].Path
	}
	originPath := strings.Split(stroagePath, ",")
	for _, value := range originPath {
		cachePaths = append(cachePaths, strings.TrimRight(value, "/")+"/"+
			e.namespace+"/"+e.name+"/jindofsx")
	}
	metaPath := cachePaths[0]
	dataPath := strings.Join(cachePaths, ",")

	var quotas []string
	var userSetQuota []string // 1Gi or 1Gi,2Gi,3Gi
	if len(runtime.Spec.TieredStore.Levels) == 0 {
		userSetQuota = append(userSetQuota, "1Gi")
		quotas = append(quotas, "1Gi")
	} else if runtime.Spec.TieredStore.Levels[0].Quota != nil {
		userSetQuota = append(userSetQuota, utils.TransformQuantityToJindoUnit(runtime.Spec.TieredStore.Levels[0].Quota))
		quotas = append(quotas, runtime.Spec.TieredStore.Levels[0].Quota.String())
	}

	if len(runtime.Spec.TieredStore.Levels) != 0 && runtime.Spec.TieredStore.Levels[0].QuotaList != "" {
		quotaList := runtime.Spec.TieredStore.Levels[0].QuotaList
		quotas = strings.Split(quotaList, ",")
		if len(quotas) != len(originPath) {
			err = fmt.Errorf("the num of cache path and quota must be equal")
			return
		}
		for _, value := range quotas {
			if strings.HasSuffix(value, "Gi") {
				value = strings.ReplaceAll(value, "Gi", "g")
			}
			if strings.HasSuffix(value, "Mi") {
				value = strings.ReplaceAll(value, "Mi", "m")
			}
			userSetQuota = append(userSetQuota, value)
		}
	}
	userQuotas := strings.Join(userSetQuota, ",") // 1g or 1g,2g

	smartdataConfig := e.getSmartDataConfigs(runtime)
	smartdataTag := smartdataConfig.imageTag
	jindoFuseImage, fuseTag, fuseImagePullPolicy := e.parseFuseImage(runtime)
	secretMountSupport := e.checkIfSupportSecretMount(runtime, smartdataTag, fuseTag)

	var mediumType = common.Memory
	var volumeType = common.VolumeTypeHostPath

	if len(runtime.Spec.TieredStore.Levels) > 0 {
		mediumType = runtime.Spec.TieredStore.Levels[0].MediumType
		volumeType = runtime.Spec.TieredStore.Levels[0].VolumeType
	}

	value = &Jindo{
		Image:               smartdataConfig.image,
		ImageTag:            smartdataConfig.imageTag,
		ImagePullPolicy:     smartdataConfig.imagePullPolicy,
		FuseImage:           jindoFuseImage,
		FuseImageTag:        fuseTag,
		FuseImagePullPolicy: fuseImagePullPolicy,
		User:                0,
		Group:               0,
		FsGroup:             0,
		UseHostNetwork:      true,
		UseHostPID:          true,
		Properties:          e.transformPriority(metaPath),
		Master: Master{
			ReplicaCount: e.transformReplicasCount(runtime),
			ServiceCount: e.transformReplicasCount(runtime),
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
			Master:            e.transformMasterMountPath(metaPath, mediumType, volumeType),
			WorkersAndClients: e.transformWorkerMountPath(originPath, quotas, e.getMediumTypeFromVolumeSource(string(mediumType), runtime.Spec.TieredStore.Levels), volumeType),
		},
		Owner: transfromer.GenerateOwnerReferenceFromObject(runtime),
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: e.namespace,
			Name:      e.name,
		},
	}
	err = e.transformHadoopConfig(runtime, value)
	if err != nil {
		return
	}
	err = e.allocatePorts(value)
	if err != nil {
		return
	}
	e.transformNetworkMode(runtime, value)
	e.transformFuseNodeSelector(runtime, value)
	e.transformSecret(runtime, value)
	err = e.transformMaster(runtime, metaPath, value, dataset, secretMountSupport)
	if err != nil {
		return
	}
	e.transformToken(runtime, value)
	e.transformWorker(runtime, dataPath, userQuotas, value)
	e.transformFuse(runtime, value)
	e.transformInitPortCheck(value)
	err = e.transformPodMetadata(runtime, value)
	if err != nil {
		return
	}
	e.transformEnvVariables(runtime, value)
	e.transformPlacementMode(dataset, value)
	e.transformRunAsUser(runtime, value)
	e.transformTolerations(dataset, runtime, value)
	err = e.transformResources(runtime, value, userQuotas)
	if err != nil {
		return
	}
	e.transformLogConfig(runtime, value)
	e.transformDeployMode(runtime, value)
	value.Master.DnsServer = smartdataConfig.dnsServer
	value.Master.NameSpace = e.namespace
	value.Fuse.MountPath = jindoFuseMountpath
	return value, err
}

func (e *JindoFSxEngine) transformMaster(runtime *datav1alpha1.JindoRuntime, metaPath string, value *Jindo, dataset *datav1alpha1.Dataset, secretMountSupport bool) (err error) {
	properties := map[string]string{
		"namespace.cluster.id":                      "local",
		"namespace.oss.copy.size":                   "1073741824",
		"namespace.filelet.threads":                 "200",
		"namespace.blocklet.threads":                "100",
		"namespace.long-running.threads":            "20",
		"namespace.filelet.cache.size":              "100000",
		"namespace.blocklet.cache.size":             "1000000",
		"namespace.filelet.atime.enable":            "false",
		"namespace.permission.root.inode.perm.bits": "511",
		"namespace.delete.scan.interval.second":     "20",
		"namespace.delete.scan.batch.size":          "5000",
		"namespace.backend.type":                    "rocksdb",
	}
	if value.Master.ReplicaCount == 3 {
		properties["namespace.backend.type"] = "raft"
		var raftLists []string
		for i := 0; i < value.Master.ReplicaCount; i++ {
			raftLists = append(raftLists, e.getMasterName()+"-"+strconv.Itoa(i)+":"+strconv.Itoa(value.Master.Port.Raft)+":0")
		}
		properties["namespace.backend.raft.initial-conf"] = strings.Join(raftLists, ",")
	}
	properties["namespace.rpc.port"] = strconv.Itoa(value.Master.Port.Rpc)
	properties["namespace.meta-dir"] = metaPath + "/server"
	// combine properties together
	if len(runtime.Spec.Master.Properties) > 0 {
		for k, v := range runtime.Spec.Master.Properties {
			properties[k] = v
		}
	}
	value.Master.MasterProperties = properties
	// to set filestore properties with confvalue
	propertiesFileStore := map[string]string{}

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

		// support nas storage
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			if len(value.UFSVolumes) == 0 {
				value.UFSVolumes = []UFSVolume{}
			}
			// Split MountPoint into PVC name and subpath (if it contains a subpath)
			parts := strings.SplitN(strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String()), "/", 2)

			if len(parts) > 1 {
				// MountPoint contains subpath
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          parts[0],
					SubPath:       parts[1],
					ContainerPath: utils.UFSPathBuilder{}.GenLocalStoragePath(mount),
				})
			} else {
				// MountPoint does not contain subpath
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          parts[0],
					ContainerPath: utils.UFSPathBuilder{}.GenLocalStoragePath(mount),
				})
			}
		} else {
			if !strings.HasSuffix(mount.MountPoint, "/") {
				mount.MountPoint = mount.MountPoint + "/"
			}
			if strings.HasPrefix(mount.MountPoint, "local:///") {
				value.Mounts.Master[mount.Name] = &Level{
					Path: mount.MountPoint[8:],
					Type: "hostPath",
				}
				value.Mounts.WorkersAndClients[mount.Name] = &Level{
					Path: mount.MountPoint[8:],
					Type: "hostPath",
				}
				continue
			}
		}

		mountType := "oss"
		if strings.HasPrefix(mount.MountPoint, "oss://") {
			var re = regexp.MustCompile(`(oss://(.*?))(/)`)
			rm := re.FindStringSubmatch(mount.MountPoint)
			if len(rm) < 3 {
				err = fmt.Errorf("incorrect oss mountPoint with %v, please check your path is dir or file ", mount.MountPoint)
				e.Log.Error(err, "mount.MountPoint", mount.MountPoint)
				return
			}
			bucketName := rm[2]
			if mount.Options["fs.oss.accessKeyId"] != "" {
				propertiesFileStore["jindofsx.oss.bucket."+bucketName+".accessKeyId"] = mount.Options["fs.oss.accessKeyId"]
			}
			if mount.Options["fs.oss.accessKeySecret"] != "" {
				propertiesFileStore["jindofsx.oss.bucket."+bucketName+".accessKeySecret"] = mount.Options["fs.oss.accessKeySecret"]
			}
			if mount.Options["fs.oss.endpoint"] == "" {
				err = fmt.Errorf("oss endpoint can not be null, please check <fs.oss.accessKeySecret> option")
				e.Log.Error(err, "oss endpoint can not be null")
				return
			}
			propertiesFileStore["jindofsx.oss.bucket."+bucketName+".endpoint"] = mount.Options["fs.oss.endpoint"]
			if strings.Contains(mount.Options["fs.oss.endpoint"], "dls") {
				propertiesFileStore["jindofsx.oss.bucket."+bucketName+".data.lake.storage.enable"] = "true"
			}
		}

		// support s3
		if strings.HasPrefix(mount.MountPoint, "s3://") {
			mountType = "s3"
			if mount.Options["fs.s3.accessKeyId"] != "" {
				propertiesFileStore["jindofsx.s3.accessKeyId"] = mount.Options["fs.s3.accessKeyId"]
			}
			if mount.Options["fs.s3.accessKeySecret"] != "" {
				propertiesFileStore["jindofsx.s3.accessKeySecret"] = mount.Options["fs.s3.accessKeySecret"]
			}
			if mount.Options["fs.s3.endpoint"] != "" {
				propertiesFileStore["jindofsx.s3.endpoint"] = mount.Options["fs.s3.endpoint"]
			}
			if mount.Options["fs.s3.region"] != "" {
				propertiesFileStore["jindofsx.s3.region"] = mount.Options["fs.s3.region"]
			}
		}

		// support cos
		if strings.HasPrefix(mount.MountPoint, "cos://") {
			mountType = "cos"
			if mount.Options["fs.cos.accessKeyId"] != "" {
				propertiesFileStore["jindofsx.cos.accessKeyId"] = mount.Options["fs.cos.accessKeyId"]
			}
			if mount.Options["fs.cos.accessKeySecret"] != "" {
				propertiesFileStore["jindofsx.cos.accessKeySecret"] = mount.Options["fs.cos.accessKeySecret"]
			}
			if mount.Options["fs.cos.endpoint"] != "" {
				propertiesFileStore["jindofsx.cos.endpoint"] = mount.Options["fs.cos.endpoint"]
			}
		}

		// support obs
		if strings.HasPrefix(mount.MountPoint, "obs://") {
			mountType = "obs"
			if mount.Options["fs.obs.accessKeyId"] != "" {
				propertiesFileStore["jindofsx.obs.accessKeyId"] = mount.Options["fs.obs.accessKeyId"]
			}
			if mount.Options["fs.obs.accessKeySecret"] != "" {
				propertiesFileStore["jindofsx.obs.accessKeySecret"] = mount.Options["fs.obs.accessKeySecret"]
			}
			if mount.Options["fs.obs.endpoint"] != "" {
				propertiesFileStore["jindofsx.obs.endpoint"] = mount.Options["fs.obs.endpoint"]
			}
		}

		// support HDFS HA
		if strings.HasPrefix(mount.MountPoint, "hdfs://") {
			mountType = "hdfs"
			for key, value := range mount.Options {
				propertiesFileStore[strings.Replace(key, "fs", "jindofsx", 1)] = value
			}
		}

		// to check whether encryptOptions exist
		for _, encryptOption := range mount.EncryptOptions {
			key := encryptOption.Name
			secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
			if secretMountSupport {
				value.Secret = secretKeyRef.Name
				if key == "fs."+mountType+".accessKeyId" {
					value.SecretKey = key
					e.Log.Info("Get %s From %s!", key, secretKeyRef.Name)
				}
				if key == "fs."+mountType+".accessKeySecret" {
					value.SecretValue = key
					e.Log.Info("Get %s From %s!", key, secretKeyRef.Name)
				}
			} else {
				secret, err := kubeclient.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
				if err != nil {
					e.Log.Info("can't get the input secret from dataset", secretKeyRef.Name)
					break
				}
				value := secret.Data[secretKeyRef.Key]
				if key == "fs."+mountType+".accessKeyId" {
					propertiesFileStore["jindofsx."+mountType+".accessKeyId"] = string(value)
				}
				if key == "fs."+mountType+".accessKeySecret" {
					propertiesFileStore["jindofsx."+mountType+".accessKeySecret"] = string(value)
				}
				e.Log.Info("Get Credential From Secret Successfully")
			}
		}
		value.MountType = mountType
	}
	value.Master.FileStoreProperties = propertiesFileStore
	return nil
}

func (e *JindoFSxEngine) transformWorker(runtime *datav1alpha1.JindoRuntime, dataPath string, userQuotas string, value *Jindo) {

	properties := map[string]string{
		"storage.cluster.id":                   "local",
		"storage.compaction.enable":            "true",
		"storage.compaction.period.minute":     "2",
		"storage.maintainence.period.minute":   "2",
		"storage.compaction.threshold":         "16",
		"storage.cache.filelet.worker.threads": "200",
		"storage.address":                      "localhost",
	}

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["storage.ram.cache.size"] = userQuotas
		//properties["storage.ram.cache.size"] = "90g"

		properties["storage.slicelet.buffer.size"] = userQuotas
		//properties["storage.slicelet.buffer.size"] = "90g"
	}

	properties["storage.rpc.port"] = strconv.Itoa(value.Worker.Port.Rpc)

	properties["storage.data-dirs"] = dataPath
	//properties["storage.data-dirs"] = "/mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot"
	if !value.UseHostNetwork {
		value.Worker.Path = dataPath
	}

	if len(runtime.Spec.TieredStore.Levels) == 0 {
		properties["storage.watermark.high.ratio"] = "0.8"
	} else {
		properties["storage.watermark.high.ratio"] = runtime.Spec.TieredStore.Levels[0].High
	}

	if len(runtime.Spec.TieredStore.Levels) == 0 {
		properties["storage.watermark.low.ratio"] = "0.6"
	} else {
		properties["storage.watermark.low.ratio"] = runtime.Spec.TieredStore.Levels[0].Low
	}

	properties["storage.data-dirs.capacities"] = userQuotas
	///properties["storage.data-dirs.capacities"] = "80g,80g,80g"

	if len(runtime.Spec.Worker.Properties) > 0 {
		for k, v := range runtime.Spec.Worker.Properties {
			properties[k] = v
		}
	}
	value.Worker.WorkerProperties = properties
}

func (e *JindoFSxEngine) transformResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
	err = e.transformMasterResources(runtime, value, userQuotas)
	if err != nil {
		return
	}
	err = e.transformWorkerResources(runtime, value, userQuotas)
	if err != nil {
		return
	}
	e.transformFuseResources(runtime, value)

	return
}

func (e *JindoFSxEngine) transformMasterResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
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

	limitMemEnable := false
	if os.Getenv("USE_DEFAULT_MEM_LIMIT") == "true" {
		limitMemEnable = true
	}
	// set memory request for the larger
	if e.hasTieredStore(runtime) && e.getTieredStoreType(runtime) == 0 {
		quotaString := strings.TrimRight(userQuotas, "g")
		needUpdated := false
		if quotaString != "" {
			i, _ := strconv.Atoi(quotaString)
			if limitMemEnable && i > defaultMemLimit {
				// value.Master.Resources.Requests.Memory = defaultMetaSize
				defaultMetaSizeQuantity := resource.MustParse(defaultMetaSize)
				if runtime.Spec.Master.Resources.Requests == nil ||
					runtime.Spec.Master.Resources.Requests.Memory() == nil ||
					runtime.Spec.Master.Resources.Requests.Memory().IsZero() ||
					defaultMetaSizeQuantity.Cmp(*runtime.Spec.Master.Resources.Requests.Memory()) > 0 {
					needUpdated = true
				}

				if !runtime.Spec.Master.Resources.Limits.Memory().IsZero() &&
					defaultMetaSizeQuantity.Cmp(*runtime.Spec.Master.Resources.Limits.Memory()) > 0 {
					return fmt.Errorf("the memory meta store's size %v is greater than master limits memory %v",
						defaultMetaSizeQuantity, runtime.Spec.Master.Resources.Limits.Memory())
				}

				if needUpdated {
					value.Master.Resources.Requests.Memory = defaultMetaSize
					err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
						runtime, err := e.getRuntime()
						if err != nil {
							return err
						}
						runtimeToUpdate := runtime.DeepCopy()
						if len(runtimeToUpdate.Spec.Master.Resources.Requests) == 0 {
							runtimeToUpdate.Spec.Master.Resources.Requests = make(corev1.ResourceList)
						}
						runtimeToUpdate.Spec.Master.Resources.Requests[corev1.ResourceMemory] = defaultMetaSizeQuantity
						if !reflect.DeepEqual(runtimeToUpdate, runtime) {
							err = e.Client.Update(context.TODO(), runtimeToUpdate)
							if err != nil {
								if apierrors.IsConflict(err) {
									time.Sleep(3 * time.Second)
								}
								return err
							}
							time.Sleep(1 * time.Second)
						}

						return nil
					})

					if err != nil {
						return err
					}

				}
			}
		}
	}

	return nil
}

func (e *JindoFSxEngine) transformWorkerResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
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

	// mem set request
	if e.hasTieredStore(runtime) && e.getTieredStoreType(runtime) == 0 {
		userQuotas = strings.ReplaceAll(userQuotas, "g", "Gi")
		needUpdated := false
		userQuotasQuantity := resource.MustParse(userQuotas)
		if runtime.Spec.Worker.Resources.Requests == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory() == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory().IsZero() ||
			userQuotasQuantity.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}

		if !runtime.Spec.Worker.Resources.Limits.Memory().IsZero() &&
			userQuotasQuantity.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the memory tierdStore's size %v is greater than master limits memory %v",
				userQuotasQuantity, runtime.Spec.Worker.Resources.Limits.Memory())
		}
		if needUpdated {
			value.Worker.Resources.Requests.Memory = userQuotas
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := e.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Worker.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Worker.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Worker.Resources.Requests[corev1.ResourceMemory] = userQuotasQuantity
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = e.Client.Update(context.TODO(), runtimeToUpdate)
					if err != nil {
						if apierrors.IsConflict(err) {
							time.Sleep(3 * time.Second)
						}
						return err
					}
					time.Sleep(1 * time.Second)
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

	}

	return
}

func (e *JindoFSxEngine) transformFuseResources(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
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
}

func (e *JindoFSxEngine) transformFuse(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	// default enable data-cache and disable meta-cache
	properties := map[string]string{
		"fs.jindofsx.request.user":           "root",
		"fs.jindofsx.data.cache.enable":      "true",
		"fs.jindofsx.meta.cache.enable":      "true",
		"fs.jindofsx.tmp.data.dir":           "/tmp",
		"fs.jindofsx.client.metrics.enable":  "true",
		"fs.oss.download.queue.size":         "16",
		"fs.oss.download.thread.concurrency": "32",
		"fs.s3.download.queue.size":          "16",
		"fs.s3.download.thread.concurrency":  "32",
	}

	readOnly := false
	runtimeInfo := e.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(e.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			e.Log.Info("Error:", "err", err)
		}
		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == corev1.ReadOnlyMany {
					readOnly = true
				}
			}
		}
	}
	// to set read only flag
	if readOnly {
		properties["fs.jindofsx.read.only.enable"] = "true"
	}

	for k, v := range value.Master.FileStoreProperties {
		// to transform jindofsx.oss.bucket to fs.jindofsx.oss.bucket
		properties[strings.Replace(k, "jindofsx", "fs", 1)] = v
	}

	// "client.storage.rpc.port": "6101",
	properties["fs.jindofsx.storage.rpc.port"] = strconv.Itoa(value.Worker.Port.Rpc)

	if e.getTieredStoreType(runtime) == 0 {
		// MEM
		properties["fs.jindofsx.ram.cache.enable"] = "true"
	} else if e.getTieredStoreType(runtime) == 1 || e.getTieredStoreType(runtime) == 2 {
		// HDD and SSD
		properties["fs.jindofsx.ram.cache.enable"] = "false"
	}
	// set secret
	if len(value.Secret) != 0 {
		properties["fs."+value.MountType+".credentials.provider"] = "com.aliyun.jindodata.oss.auth.CustomCredentialsProvider"
		properties["aliyun."+value.MountType+".provider.url"] = "secrets:///token/"
		properties["fs."+value.MountType+".provider.endpoint"] = "secrets:///token/"
		properties["fs."+value.MountType+".provider.format"] = "JSON"
	}

	if len(runtime.Spec.Fuse.Properties) > 0 {
		for k, v := range runtime.Spec.Fuse.Properties {
			properties[k] = v
		}
	}
	value.Fuse.FuseProperties = properties

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()
}

func (e *JindoFSxEngine) transformLogConfig(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	fsxProperties := map[string]string{
		"application.report.on":  "true",
		"metric.report.on":       "true",
		"logger.dir":             "/tmp/jindofsx-log",
		"logger.cleanner.enable": "true",
		"logger.consolelogger":   "true",
		"logger.jnilogger":       "false",
		"logger.sync":            "false",
		"logger.verbose":         "0",
	}

	fusePropreties := map[string]string{
		"logger.dir":            "/tmp/fuse-log",
		"logger.consolelogger":  "true",
		"logger.level":          "2",
		"logger.cleaner.enable": "true",
		"logger.sync":           "false",
		"logger.verbose":        "0",
	}

	if len(runtime.Spec.LogConfig) > 0 {
		for k, v := range runtime.Spec.LogConfig {
			fsxProperties[k] = v
		}
	}

	if len(runtime.Spec.Fuse.LogConfig) > 0 {
		for k, v := range runtime.Spec.Fuse.LogConfig {
			fusePropreties[k] = v
		}
	}

	value.LogConfig = fsxProperties
	value.FuseLogConfig = fusePropreties
}

func (e *JindoFSxEngine) transformFuseNodeSelector(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}

	// The label will be added by CSI Plugin when any workload pod is scheduled on the node.
	value.Fuse.NodeSelector[e.getFuseLabelname()] = "true"
}

func (e *JindoFSxEngine) transformNodeSelector(runtime *datav1alpha1.JindoRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Worker.NodeSelector != nil {
		properties = runtime.Spec.Worker.NodeSelector
	}
	return properties
}

func (e *JindoFSxEngine) transformReplicasCount(runtime *datav1alpha1.JindoRuntime) int {
	if runtime.Spec.Master.Replicas == JINDO_HA_MASTERNUM {
		return JINDO_HA_MASTERNUM
	}
	return JINDO_MASTERNUM_DEFAULT
}

func (e *JindoFSxEngine) transformMasterSelector(runtime *datav1alpha1.JindoRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Master.NodeSelector != nil {
		properties = runtime.Spec.Master.NodeSelector
	}
	return properties
}

func (e *JindoFSxEngine) transformPriority(metaPath string) map[string]string {
	properties := map[string]string{}
	properties["logDir"] = metaPath + "/log"
	return properties
}

func (e *JindoFSxEngine) transformMasterMountPath(metaPath string, mediumType common.MediumType, volumeType common.VolumeType) map[string]*Level {
	properties := map[string]*Level{}
	properties["1"] = &Level{
		Path:       metaPath,
		Type:       string(volumeType),
		MediumType: string(mediumType),
	}
	return properties
}

func (e *JindoFSxEngine) transformWorkerMountPath(originPath []string, quotas []string, mediumType string, volumeType common.VolumeType) map[string]*Level {
	properties := map[string]*Level{}
	for index, value := range originPath {
		mountVol := &Level{
			Path:       strings.TrimRight(value, "/"),
			Type:       string(volumeType),
			MediumType: mediumType,
			Quota:      quotas[index],
		}
		//properties[strconv.Itoa(index+1)] = strings.TrimRight(value, "/")
		properties[strconv.Itoa(index+1)] = mountVol
	}
	return properties
}

func (e *JindoFSxEngine) transformFuseArg(runtime *datav1alpha1.JindoRuntime, dataset *datav1alpha1.Dataset) []string {
	fuseArgs := []string{}
	readOnly := false
	runtimeInfo := e.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(e.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			e.Log.Info("Error:", "err", err)
		}
		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == corev1.ReadOnlyMany {
					readOnly = true
				}
			}
		}
	}
	if len(runtime.Spec.Fuse.Args) > 0 {
		fuseArgs = runtime.Spec.Fuse.Args
	} else {
		if readOnly {
			fuseArgs = append(fuseArgs, "-okernel_cache")
			fuseArgs = append(fuseArgs, "-oro")
			fuseArgs = append(fuseArgs, "-oattr_timeout=7200")
			fuseArgs = append(fuseArgs, "-oentry_timeout=7200")
			fuseArgs = append(fuseArgs, "-onegative_timeout=7200")
		} else {
			fuseArgs = append(fuseArgs, "-oauto_cache")
			fuseArgs = append(fuseArgs, "-oattr_timeout=0")
			fuseArgs = append(fuseArgs, "-oentry_timeout=0")
			fuseArgs = append(fuseArgs, "-onegative_timeout=0")
		}
		fuseArgs = append(fuseArgs, "-ometrics_port=0")
	}
	if runtime.Spec.Master.Disabled && runtime.Spec.Worker.Disabled {
		fuseArgs = append(fuseArgs, "-ouri="+dataset.Spec.Mounts[0].MountPoint)
	}
	return fuseArgs
}

func (e *JindoFSxEngine) getSmartDataConfigs(runtime *datav1alpha1.JindoRuntime) smartdataConfig {
	// Apply defaults
	config := smartdataConfig{
		image:           "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata",
		imageTag:        "4.6.8",
		imagePullPolicy: "Always",
		dnsServer:       "1.1.1.1",
	}

	// Override with global-scoped configs
	globalImage := docker.GetImageRepoFromEnv(common.JindoSmartDataImageEnv)
	globalTag := docker.GetImageTagFromEnv(common.JindoSmartDataImageEnv)
	globalDnsServer := os.Getenv(common.JindoDnsServer)
	if len(globalImage) > 0 {
		config.image = globalImage
	}
	if len(globalTag) > 0 {
		config.imageTag = globalTag
	}
	if len(globalDnsServer) > 0 {
		config.dnsServer = globalDnsServer
	}

	// Override with runtime-scoped configs
	if len(runtime.Spec.JindoVersion.Image) > 0 {
		config.image = runtime.Spec.JindoVersion.Image
	}
	if len(runtime.Spec.JindoVersion.ImageTag) > 0 {
		config.imageTag = runtime.Spec.JindoVersion.ImageTag
	}
	if len(runtime.Spec.JindoVersion.ImagePullPolicy) > 0 {
		config.imagePullPolicy = runtime.Spec.JindoVersion.ImagePullPolicy
	}

	e.Log.Info("Set image", "config", config)

	return config
}

func (e *JindoFSxEngine) parseFuseImage(runtime *datav1alpha1.JindoRuntime) (image, tag, imagePullPolicy string) {
	// Apply defaults
	image = "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse"
	tag = "4.6.8"
	imagePullPolicy = "Always"

	// Override with global-scoped configs
	globalImage := docker.GetImageRepoFromEnv(common.JindoFuseImageEnv)
	globalTag := docker.GetImageTagFromEnv(common.JindoFuseImageEnv)
	if len(globalImage) > 0 {
		image = globalImage
	}
	if len(globalTag) > 0 {
		tag = globalTag
	}

	// Override with runtime-scoped configs
	if len(runtime.Spec.Fuse.Image) > 0 {
		image = runtime.Spec.Fuse.Image
	}
	if len(runtime.Spec.Fuse.ImageTag) > 0 {
		tag = runtime.Spec.Fuse.ImageTag
	}
	if len(runtime.Spec.Fuse.ImagePullPolicy) > 0 {
		imagePullPolicy = runtime.Spec.Fuse.ImagePullPolicy
	}

	e.Log.Info("Set fuse image", "image", image, "tag", tag, "imagePullPolicy", imagePullPolicy)

	return
}

func (e *JindoFSxEngine) transformSecret(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	if len(runtime.Spec.Secret) != 0 {
		value.Secret = runtime.Spec.Secret
		value.UseStsToken = true
	} else {
		value.UseStsToken = false
	}
}

func (e *JindoFSxEngine) transformToken(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	properties := map[string]string{}
	if len(value.Secret) != 0 {
		properties["default.credential.provider"] = "secrets:///token/"
		properties["jindofsx."+value.MountType+".provider.endpoint"] = "secrets:///token/"
	} else {
		properties["default.credential.provider"] = "none"
	}
	value.Master.TokenProperties = properties
}

func (e *JindoFSxEngine) allocatePorts(value *Jindo) error {

	// if not usehostnetwork then use default port
	// usehostnetwork to choose port from port allocator
	expectedPortNum := 2
	if !value.UseHostNetwork {
		value.Master.Port.Rpc = DEFAULT_MASTER_RPC_PORT
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

func (e *JindoFSxEngine) transformInitPortCheck(value *Jindo) {
	// This function should be called after port allocation

	if !common.PortCheckEnabled() {
		return
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
}

func (e *JindoFSxEngine) transformRunAsUser(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	if len(runtime.Spec.User) != 0 {
		value.Fuse.RunAs = runtime.Spec.User
	}
}

func (e *JindoFSxEngine) transformTolerations(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.JindoRuntime, value *Jindo) {

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

func (e *JindoFSxEngine) transformPodMetadata(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	// check if setting labels with deprecated API(i.e. spec.labels)
	deprecatedLabelSet := len(runtime.Spec.Labels) != 0 ||
		len(runtime.Spec.Master.Labels) != 0 ||
		len(runtime.Spec.Worker.Labels) != 0 ||
		len(runtime.Spec.Fuse.Labels) != 0

	labelSet := len(runtime.Spec.PodMetadata.Labels) != 0 ||
		len(runtime.Spec.Master.PodMetadata.Labels) != 0 ||
		len(runtime.Spec.Worker.PodMetadata.Labels) != 0 ||
		len(runtime.Spec.Fuse.PodMetadata.Labels) != 0

	if deprecatedLabelSet && labelSet {
		return errors.New("cannot setting runtime pod's labels with both spec.labels(deprecated) and spec.podMetadata.labels. Use spec.podMetadata.labels only")
	}

	// transform labels
	if deprecatedLabelSet {
		commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.Labels)
		value.Master.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Master.Labels)
		value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.Labels)
		value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.Labels)
	} else if labelSet {
		commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Labels)
		value.Master.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Master.PodMetadata.Labels)
		value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.PodMetadata.Labels)
		value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.PodMetadata.Labels)
	}

	commonAnnotations := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Annotations)
	value.Master.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Master.PodMetadata.Annotations)
	value.Worker.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Worker.PodMetadata.Annotations)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Fuse.PodMetadata.Annotations)

	return nil
}

func (e *JindoFSxEngine) transformNetworkMode(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
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

func (e *JindoFSxEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *Jindo) {

	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (e *JindoFSxEngine) transformDeployMode(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	// transform master disabled
	if runtime.Spec.Master.Disabled {
		value.Master.ReplicaCount = 0
	}
	if runtime.Spec.Worker.Disabled {
		value.Worker.ReplicaCount = 0
	}
	// to set fuseOnly
	if runtime.Spec.Master.Disabled && runtime.Spec.Worker.Disabled {
		value.Fuse.Mode = FuseOnly
	}
}

func (e *JindoFSxEngine) transformEnvVariables(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	if len(runtime.Spec.Master.Env) > 0 {
		value.Master.Env = runtime.Spec.Master.Env
	}

	if len(runtime.Spec.Worker.Env) > 0 {
		value.Worker.Env = runtime.Spec.Worker.Env
	}

	if len(runtime.Spec.Fuse.Env) > 0 {
		value.Fuse.Env = runtime.Spec.Fuse.Env
	}
}

func (e *JindoFSxEngine) getMediumTypeFromVolumeSource(defaultMediumType string, levels []datav1alpha1.Level) string {
	var mediumType = defaultMediumType

	if len(levels) > 0 {
		if levels[0].VolumeType == common.VolumeTypeEmptyDir {
			if levels[0].VolumeSource.EmptyDir != nil {
				mediumType = string(levels[0].VolumeSource.EmptyDir.Medium)
			}
		}
	}

	return mediumType
}

func (e *JindoFSxEngine) checkIfSupportSecretMount(runtime *datav1alpha1.JindoRuntime, smartdataTag string, fuseTag string) bool {
	fuseOnly := runtime.Spec.Master.Disabled && runtime.Spec.Worker.Disabled
	compareSmartdata, _ := versionutil.Compare(smartdataTag, imageTagSupportAKFile)
	newSmartdataVersion := compareSmartdata >= 0
	compareFuse, _ := versionutil.Compare(fuseTag, imageTagSupportAKFile)
	newFuseVersion := compareFuse >= 0
	if (fuseOnly && newFuseVersion) || (newSmartdataVersion && newFuseVersion) {
		return true
	}
	return false
}
