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

package alluxio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
)

func (e *AlluxioEngine) transform(runtime *datav1alpha1.AlluxioRuntime) (value *Alluxio, err error) {
	if runtime == nil {
		err = fmt.Errorf("the alluxioRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "AlluxioRuntime.Transform", "name", runtime.Name)

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return value, err
	}

	value = &Alluxio{
		Owner: transfromer.GenerateOwnerReferenceFromObject(runtime),
	}

	value.FullnameOverride = e.name

	// 1.transform the common part
	err = e.transformCommonPart(runtime, dataset, value)
	if err != nil {
		return
	}

	// 2.transform the masters
	err = e.transformMasters(runtime, dataset, value)
	if err != nil {
		return
	}

	// 3.transform the workers
	err = e.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	// 4.transform the fuse
	err = e.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	err = e.transformPodMetadata(runtime, value)
	if err != nil {
		return
	}

	// 5.transform the hadoop non-default configurations
	err = e.transformHadoopConfig(runtime, value)
	if err != nil {
		return
	}

	// 6.transform the dataset if it has local path or volume
	e.transformDatasetToVolume(runtime, dataset, value)

	// 7.transform the permission
	e.transformPermission(runtime, value)

	// 8.set optimization parameters
	e.optimizeDefaultProperties(runtime, value)

	// 9. set optimization parameters if all the mounts are HTTP
	e.optimizeDefaultPropertiesAndFuseForHTTP(runtime, dataset, value)

	// 10.allocate port for fluid engine
	if datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode) ||
		datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode) {
		e.Log.Info("allocatePorts for hostnetwork mode")
		err = e.allocatePorts(value, runtime)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip allocatePorts for container network mode")
		e.generateStaticPorts(value)
	}

	// 11.set engine properties
	e.setPortProperties(runtime, value)

	// 12.set API Gateway
	err = e.transformAPIGateway(runtime, value)

	// 13.set the placementMode
	e.transformPlacementMode(dataset, value)
	return
}

// 2. Transform the common part
func (e *AlluxioEngine) transformCommonPart(runtime *datav1alpha1.AlluxioRuntime,
	dataset *datav1alpha1.Dataset,
	value *Alluxio,
) (err error) {
	value.RuntimeIdentity.Namespace = runtime.Namespace
	value.RuntimeIdentity.Name = runtime.Name

	image := runtime.Spec.AlluxioVersion.Image
	imageTag := runtime.Spec.AlluxioVersion.ImageTag
	imagePullPolicy := runtime.Spec.AlluxioVersion.ImagePullPolicy

	// TODO: support imagePullSecrets by AlluxioRuntime
	imagePullSecrets := []corev1.LocalObjectReference{}

	value.Image, value.ImageTag, value.ImagePullPolicy, value.ImagePullSecrets, err = e.parseRuntimeImage(image, imageTag, imagePullPolicy, imagePullSecrets)
	if err != nil {
		return err
	}

	value.UserInfo = common.UserInfo{
		User: 0,
		// FSGroup: 0,
		Group: 0,
	}

	// transform init users
	e.transformInitUsers(runtime, value)

	// TODO: support nodeAffinity

	if len(runtime.Spec.Properties) > 0 {
		value.Properties = runtime.Spec.Properties
	} else {
		value.Properties = map[string]string{}
	}

	// generate alluxio root ufs by dataset spec mounts
	e.Log.Info("input", "mounts", dataset.Spec.Mounts, "common.RootDirPath", common.RootDirPath)
	uRootPath, m := utils.UFSPathBuilder{}.GenAlluxioUFSRootPath(dataset.Spec.Mounts)
	// attach mount options when direct mount ufs endpoint
	if m != nil {
		extractEncryptOption := !IsMountWithConfigMap()
		if mOptions, err := e.genUFSMountOptions(*m, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions, extractEncryptOption); err != nil {
			return err
		} else {
			for k, v := range mOptions {
				value.Properties[k] = v
			}
		}
	}
	e.Log.Info("output", "uRootPath", uRootPath, "m", m)
	// set alluxio root ufs
	value.Properties["alluxio.master.mount.table.root.ufs"] = uRootPath

	// Set the max replication
	dataReplicas := runtime.Spec.Data.Replicas
	if dataReplicas <= 0 {
		dataReplicas = 1
	}
	value.Properties["alluxio.user.file.replication.max"] = fmt.Sprintf("%d", dataReplicas)

	if len(runtime.Spec.JvmOptions) > 0 {
		value.JvmOptions = runtime.Spec.JvmOptions
	}

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	// Set tieredstore levels
	var levels []Level
	for _, level := range runtimeInfo.GetTieredStoreInfo().Levels {

		mediumType := e.getMediumTypeFromVolumeSource(string(level.MediumType), level)
		l := tieredstore.GetTieredLevel(runtimeInfo, level.MediumType)

		var paths []string
		var quotas []string
		for _, cachePath := range level.CachePaths {
			paths = append(paths, fmt.Sprintf("%s/%s/%s", cachePath.Path, runtime.Namespace, runtime.Name))
			quotas = append(quotas, utils.TransformQuantityToAlluxioUnit(cachePath.Quota))
		}

		pathConfigStr := strings.Join(paths, ",")
		quotaConfigStr := strings.Join(quotas, ",")
		mediumTypeConfigStr := strings.Join(*utils.FillSliceWithString(mediumType, len(paths)), ",")

		levels = append(levels, Level{
			Alias:      string(level.MediumType),
			Level:      l,
			Type:       string(level.VolumeType),
			Path:       pathConfigStr,
			MediumType: mediumTypeConfigStr,
			Low:        level.Low,
			High:       level.High,
			Quota:      quotaConfigStr,
		})
	}

	value.TieredStore.Levels = levels

	// value.Metastore = Metastore{
	// 	VolumeType: "emptyDir",
	// }

	// quantity, err := resource.ParseQuantity("10Gi")
	// if err != nil {
	// 	return err
	// }

	value.Journal = Journal{
		VolumeType: "emptyDir",
		Size:       "30Gi",
	}

	if !runtime.Spec.DisablePrometheus {
		value.Monitoring = alluxioRuntimeMetricsLabel
	}

	// transform Tolerations
	e.transformTolerations(dataset, value)

	e.transformShortCircuit(runtimeInfo, value)

	return
}

func (e *AlluxioEngine) transformPodMetadata(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	// transform labels
	commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Labels)
	value.Master.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Master.PodMetadata.Labels)
	value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.PodMetadata.Labels)
	value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.PodMetadata.Labels)

	// transform annotations
	commonAnnotations := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Annotations)
	value.Master.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Master.PodMetadata.Annotations)
	value.Worker.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Worker.PodMetadata.Annotations)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Fuse.PodMetadata.Annotations)

	return nil
}

// 2. Transform the masters
func (e *AlluxioEngine) transformMasters(runtime *datav1alpha1.AlluxioRuntime,
	dataset *datav1alpha1.Dataset,
	value *Alluxio,
) (err error) {
	value.Master = Master{}

	backupRoot := os.Getenv("FLUID_WORKDIR")
	if backupRoot == "" {
		backupRoot = "/tmp"
	}
	value.Master.BackupPath = backupRoot + "/alluxio-backup/" + e.namespace + "/" + e.name

	if runtime.Spec.Master.Replicas == 0 {
		value.Master.Replicas = 1
	} else {
		value.Master.Replicas = runtime.Spec.Master.Replicas
	}

	// if len(runtime.Spec.Master.JvmOptions) > 0 {
	// 	value.Master.JvmOptions = strings.Join(runtime.Spec.Master.JvmOptions, " ")
	// }

	e.optimizeDefaultForMaster(runtime, value)

	if len(runtime.Spec.Master.Env) > 0 {
		value.Master.Env = runtime.Spec.Master.Env
	} else {
		value.Master.Env = map[string]string{}
	}

	value.Master.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path(e.name, e.namespace)

	if len(runtime.Spec.Master.Properties) > 0 {
		value.Master.Properties = runtime.Spec.Master.Properties
		runtime.Spec.Properties = utils.UnionMapsWithOverride(runtime.Spec.Properties, runtime.Spec.Master.Properties)
	}

	// parse master pod network mode
	value.Master.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode)

	nodeSelector := e.transformMasterSelector(runtime)
	if len(nodeSelector) != 0 {
		value.Master.NodeSelector = nodeSelector
	}

	// // check the run as
	// if runtime.Spec.RunAs != nil {
	// 	value.Master.Env["ALLUXIO_USERNAME"] = alluxioUser
	// 	value.Master.Env["ALLUXIO_GROUP"] = alluxioUser
	// 	value.Master.Env["ALLUXIO_UID"] = strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	// 	value.Master.Env["ALLUXIO_GID"] = strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	// }
	// if the dataset indicates a restore path, need to load the  backup file in it

	if dataset.Spec.DataRestoreLocation != nil {
		if dataset.Spec.DataRestoreLocation.Path != "" {
			pvcName, path, err := utils.ParseBackupRestorePath(dataset.Spec.DataRestoreLocation.Path)
			if err != nil {
				e.Log.Error(err, "restore path cannot analyse", "Path", dataset.Spec.DataRestoreLocation.Path)
			}
			if pvcName != "" {
				// RestorePath is in the form of pvc://<pvcName>/subpath
				value.Master.Restore.Enabled = true
				value.Master.Restore.PVCName = pvcName
				value.Master.Restore.Path = path
				value.Master.Env["JOURNAL_BACKUP"] = "/pvc" + path + e.GetMetadataFileName()
			} else if dataset.Spec.DataRestoreLocation.NodeName != "" {
				// RestorePath is in the form of local://subpath
				value.Master.Restore.Enabled = true
				if len(value.Master.NodeSelector) == 0 {
					value.Master.NodeSelector = map[string]string{}
				}
				value.Master.NodeSelector["kubernetes.io/hostname"] = dataset.Spec.DataRestoreLocation.NodeName
				value.Master.Env["JOURNAL_BACKUP"] = "/host/" + e.GetMetadataFileName()
				value.Master.Restore.Path = path
			} else {
				// RestorePath in Dataset cannot analyse
				err := errors.New("DataRestoreLocation in Dataset cannot analyse, will not restore")
				e.Log.Error(err, "restore path cannot analyse", "Location", dataset.Spec.DataRestoreLocation)
			}
		}
	}

	e.transformResourcesForMaster(runtime, value)

	// transform volumes for master
	err = e.transformMasterVolumes(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform volumes for master")
	}

	if IsMountWithConfigMap() {
		e.Log.Info("use configmap to generate mount info")
		// transform mount secret options
		nonNativeMounts, err := e.generateNonNativeMountsInfo(dataset)
		if err != nil {
			e.Log.Error(err, "generate non native mount info occurs error")
			return err
		}
		value.Master.NonNativeMounts = nonNativeMounts
		value.Master.MountConfigStorage = ConfigmapStorageName
		e.transformEncryptOptionsToMasterVolumes(dataset, value)
	}
	return
}

// 3. Transform the workers
func (e *AlluxioEngine) transformWorkers(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	value.Worker = Worker{}
	e.optimizeDefaultForWorker(runtime, value)

	if len(runtime.Spec.Worker.NodeSelector) > 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	} else {
		value.Worker.NodeSelector = map[string]string{}
	}

	if len(runtime.Spec.Worker.Properties) > 0 {
		value.Worker.Properties = runtime.Spec.Worker.Properties
		runtime.Spec.Properties = utils.UnionMapsWithOverride(runtime.Spec.Properties, runtime.Spec.Worker.Properties)
	}

	if len(runtime.Spec.Worker.Env) > 0 {
		value.Worker.Env = runtime.Spec.Worker.Env
	} else {
		value.Worker.Env = map[string]string{}
	}

	// check the run as
	// if runtime.Spec.RunAs != nil {
	// 	value.Worker.Env["ALLUXIO_USERNAME"] = alluxioUser
	// 	value.Worker.Env["ALLUXIO_GROUP"] = alluxioUser
	// 	value.Worker.Env["ALLUXIO_UID"] = strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	// 	value.Worker.Env["ALLUXIO_GID"] = strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	// }

	value.Worker.Env["ALLUXIO_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH"] = value.getTiredStoreLevel0Path(e.name, e.namespace)

	// parse work pod network mode
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode)

	err = e.transformResourcesForWorker(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform resource for worker")
		return err
	}

	// transform volumes for worker
	err = e.transformWorkerVolumes(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform volumes for worker")
	}

	return
}

func (e *AlluxioEngine) generateStaticPorts(value *Alluxio) {
	value.Master.Ports.Rpc = 19998
	value.Master.Ports.Web = 19999
	value.Worker.Ports.Rpc = 29999
	value.Worker.Ports.Web = 30000
	value.JobMaster.Ports.Rpc = 20001
	value.JobMaster.Ports.Web = 20002
	value.JobWorker.Ports.Rpc = 30001
	value.JobWorker.Ports.Web = 30003
	value.JobWorker.Ports.Data = 30002
	if e.runtime.Spec.APIGateway.Enabled {
		value.APIGateway.Ports.Rest = 39999
	}
	if e.runtime.Spec.Master.Replicas > 1 {
		value.Master.Ports.Embedded = 19200
		value.JobMaster.Ports.Embedded = 20003
	}
}

// 8.allocate port for fluid engine
func (e *AlluxioEngine) allocatePorts(value *Alluxio, runtime *datav1alpha1.AlluxioRuntime) error {
	expectedPortNum := portNum

	if e.runtime.Spec.APIGateway.Enabled {
		expectedPortNum += 1
	}

	if e.runtime.Spec.Master.Replicas > 1 {
		expectedPortNum += 2
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

	value.Master.Ports.Rpc, index = e.allocateSinglePort(value, "alluxio.master.rpc.port", allocatedPorts, index, runtime.Spec.Master.Ports, "rpc")
	value.Master.Ports.Web, index = e.allocateSinglePort(value, "alluxio.master.web.port", allocatedPorts, index, runtime.Spec.Master.Ports, "web")
	value.Worker.Ports.Rpc, index = e.allocateSinglePort(value, "alluxio.worker.rpc.port", allocatedPorts, index, runtime.Spec.Worker.Ports, "rpc")
	value.Worker.Ports.Web, index = e.allocateSinglePort(value, "alluxio.worker.web.port", allocatedPorts, index, runtime.Spec.Worker.Ports, "web")
	value.JobMaster.Ports.Rpc, index = e.allocateSinglePort(value, "alluxio.job.master.rpc.port", allocatedPorts, index, runtime.Spec.JobMaster.Ports, "rpc")
	value.JobMaster.Ports.Web, index = e.allocateSinglePort(value, "alluxio.job.master.web.port", allocatedPorts, index, runtime.Spec.JobMaster.Ports, "web")
	value.JobWorker.Ports.Rpc, index = e.allocateSinglePort(value, "alluxio.job.worker.rpc.port", allocatedPorts, index, runtime.Spec.JobWorker.Ports, "rpc")
	value.JobWorker.Ports.Web, index = e.allocateSinglePort(value, "alluxio.job.worker.web.port", allocatedPorts, index, runtime.Spec.JobWorker.Ports, "web")
	value.JobWorker.Ports.Data, index = e.allocateSinglePort(value, "alluxio.job.worker.data.port", allocatedPorts, index, runtime.Spec.JobWorker.Ports, "data")

	if e.runtime.Spec.APIGateway.Enabled {
		value.APIGateway.Ports.Rest, index = e.allocateSinglePort(value, "alluxio.proxy.web.port", allocatedPorts, index, runtime.Spec.APIGateway.Ports, "web")
	}

	if e.runtime.Spec.Master.Replicas > 1 {
		value.Master.Ports.Embedded, index = e.allocateSinglePort(value, "alluxio.master.embedded.journal.port", allocatedPorts, index, runtime.Spec.Master.Ports, "embeddedJournal")
		value.JobMaster.Ports.Embedded, _ = e.allocateSinglePort(value, "alluxio.job.master.embedded.journal.port", allocatedPorts, index, runtime.Spec.JobMaster.Ports, "embeddedJournal")
	}

	return nil
}

func (e *AlluxioEngine) allocateSinglePort(value *Alluxio, key string, allocatedPorts []int, index int, runtimeValue map[string]int, runtimeKey string) (newPort, newIndex int) {
	var port int
	if newVal, found := value.Properties[key]; found {
		port, _ = strconv.Atoi(newVal)
	} else if runtimeVal, found := runtimeValue[runtimeKey]; found {
		port = runtimeVal
	} else {
		port = allocatedPorts[index]
		index++
	}

	return port, index
}

func (e *AlluxioEngine) transformMasterSelector(runtime *datav1alpha1.AlluxioRuntime) map[string]string {
	properties := map[string]string{}
	if runtime.Spec.Master.NodeSelector != nil {
		properties = runtime.Spec.Master.NodeSelector
	}
	return properties
}

func (e *AlluxioEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *Alluxio) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (e *AlluxioEngine) transformShortCircuit(runtimeInfo base.RuntimeInfoInterface, value *Alluxio) {
	value.Fuse.ShortCircuitPolicy = "local"

	enableShortCircuit := true

	// Disable short circuit when using emptyDir as the volume type of any tieredstore level.
	for _, level := range runtimeInfo.GetTieredStoreInfo().Levels {
		if level.VolumeType == common.VolumeTypeEmptyDir {
			enableShortCircuit = false
			break
		}
	}

	value.ShortCircuit = ShortCircuit{
		VolumeType: "emptyDir",
		Policy:     "local",
		Enable:     enableShortCircuit,
	}

	if !enableShortCircuit {
		value.Properties["alluxio.user.short.circuit.enabled"] = "false"
	}
}

func (e *AlluxioEngine) generateNonNativeMountsInfo(dataset *datav1alpha1.Dataset) ([]string, error) {
	var nonNativeMounts []string
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		mOptions, err := e.genUFSMountOptions(mount, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions, false)
		if err != nil {
			return nil, err
		}

		alluxioPath := utils.UFSPathBuilder{}.GenAlluxioMountPath(mount)

		var mountArgs []string

		// ensure the first two element is the alluxio mount point and ufs path, used in mount.sh
		mountArgs = append(mountArgs, alluxioPath, mount.MountPoint)

		if mount.ReadOnly {
			mountArgs = append(mountArgs, "--readonly")
		}

		if mount.Shared {
			mountArgs = append(mountArgs, "--shared")
		}

		for k, v := range mOptions {
			mountArgs = append(mountArgs, "--option", fmt.Sprintf("%s=%s", k, v))
		}

		// use space as seperator as it will append to `alluxio fs mount` directly.
		nonNativeMounts = append(nonNativeMounts, strings.Join(mountArgs, " "))
	}
	return nonNativeMounts, nil
}

func (e *AlluxioEngine) getMediumTypeFromVolumeSource(defaultMediumType string, level base.Level) string {
	mediumType := defaultMediumType

	if level.VolumeType == common.VolumeTypeEmptyDir {
		if level.VolumeSource.EmptyDir != nil {
			mediumType = string(level.VolumeSource.EmptyDir.Medium)
		}
	}

	return mediumType
}
