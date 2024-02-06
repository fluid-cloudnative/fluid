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

package juicefs

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
)

func (j *JuiceFSEngine) CreateDataMigrateJob(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (err error) {
	log := ctx.Log.WithName("createDataMigrateJob")

	// 1. Check if the helm release already exists
	releaseName := utils.GetDataMigrateReleaseName(targetDataMigrate.Name)
	jobName := utils.GetDataMigrateJobName(releaseName)
	var existed bool
	existed, err = helm.CheckRelease(releaseName, targetDataMigrate.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", targetDataMigrate.Namespace)
		return err
	}

	// 2. install the helm chart if not exists
	if !existed {
		log.Info("DataMigrate job helm chart not installed yet, will install")
		valueFileName, err := j.generateDataMigrateValueFile(ctx, &targetDataMigrate)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdatamigrate.DataMigrateChart + "/" + common.JuiceFSRuntime
		err = helm.InstallRelease(releaseName, targetDataMigrate.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install datamigrate chart")
			return err
		}
		log.Info("DataLoad job helm chart successfully installed", "namespace", targetDataMigrate.Namespace, "releaseName", releaseName)
		ctx.Recorder.Eventf(&targetDataMigrate, corev1.EventTypeNormal, common.DataLoadJobStarted, "The DataMigrate job(or cronjob) %s started", jobName)
	}
	return err
}

func (j *JuiceFSEngine) generateDataMigrateValueFile(r cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataMigrate, ok := object.(*datav1alpha1.DataMigrate)
	if !ok {
		return "", fmt.Errorf("object %v is not a DataMigrate", object)
	}

	// 1. get the target dataset
	targetDataset, err := utils.GetTargetDatasetOfMigrate(r.Client, dataMigrate)
	if err != nil {
		return "", err
	}
	j.Log.Info("target dataset", "dataset", targetDataset)

	// 2. get info
	imageName, imageTag := dataMigrate.Spec.Image, dataMigrate.Spec.ImageTag

	var defaultJuiceFSImage string
	if len(imageName) == 0 || len(imageTag) == 0 {
		defaultJuiceFSImage = common.DefaultCEImage
		edition := j.GetEdition()
		if edition == EnterpriseEdition {
			defaultJuiceFSImage = common.DefaultEEImage
		}
	}

	if len(imageName) == 0 {
		defaultImageInfo := strings.Split(defaultJuiceFSImage, ":")
		if len(defaultImageInfo) < 1 {
			err = fmt.Errorf("invalid default datamigrate image")
			return
		} else {
			imageName = defaultImageInfo[0]
		}
	}

	if len(imageTag) == 0 {
		defaultImageInfo := strings.Split(defaultJuiceFSImage, ":")
		if len(defaultImageInfo) < 2 {
			err = fmt.Errorf("invalid default datamigrate image")
			return
		} else {
			imageTag = defaultImageInfo[1]
		}
	}

	image := fmt.Sprintf("%s:%s", imageName, imageTag)
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

	// 3. init dataMigrateInfo
	dataMigrateInfo := cdatamigrate.DataMigrateInfo{
		BackoffLimit:     3,
		TargetDataset:    targetDataset.Name,
		EncryptOptions:   []datav1alpha1.EncryptOption{},
		Image:            image,
		Options:          map[string]string{},
		Labels:           dataMigrate.Spec.PodMetadata.Labels,
		Annotations:      dataMigrate.Spec.PodMetadata.Annotations,
		ImagePullSecrets: imagePullSecrets,
		Policy:           string(dataMigrate.Spec.Policy),
		Schedule:         dataMigrate.Spec.Schedule,
		Resources:        dataMigrate.Spec.Resources,
		Parallelism:      dataMigrate.Spec.Parallelism,
	}

	// first set the affinity, below code will add another affinity terms.
	if dataMigrate.Spec.Affinity != nil {
		dataMigrateInfo.Affinity = dataMigrate.Spec.Affinity
	}

	// generate ssh config for parallel tasks when using parallel tasks
	if dataMigrateInfo.Parallelism > 1 {
		err = j.setParallelMigrateOptions(&dataMigrateInfo, dataMigrate)
		if err != nil {
			return "", err
		}
		// the launcher prefers to run on different host with the workers
		addWorkerPodAntiAffinity(&dataMigrateInfo, dataMigrate)
	}

	if dataMigrate.Spec.NodeSelector != nil {
		dataMigrateInfo.NodeSelector = dataMigrate.Spec.NodeSelector
	}

	if len(dataMigrate.Spec.Tolerations) > 0 {
		dataMigrateInfo.Tolerations = dataMigrate.Spec.Tolerations
	}

	if len(dataMigrate.Spec.SchedulerName) > 0 {
		dataMigrateInfo.SchedulerName = dataMigrate.Spec.SchedulerName
	}

	// 4. set options
	timeout := dataMigrate.Spec.Options["timeout"]
	delete(dataMigrate.Spec.Options, "timeout")
	if timeout == "" {
		timeout = DefaultDataMigrateTimeout
	}
	options := []string{}
	for k, v := range dataMigrate.Spec.Options {
		if v != "" {
			options = append(options, fmt.Sprintf("--%s=%s", k, v))
		} else {
			options = append(options, fmt.Sprintf("--%s", k))
		}
	}
	dataMigrateInfo.Options["option"] = strings.Join(options, " ")
	dataMigrateInfo.Options["timeout"] = timeout

	// 5. set from & to
	migrateFrom, err := j.genDataUrl(dataMigrate.Spec.From, targetDataset, &dataMigrateInfo)
	if err != nil {
		return "", err
	}
	migrateTo, err := j.genDataUrl(dataMigrate.Spec.To, targetDataset, &dataMigrateInfo)
	if err != nil {
		return "", err
	}
	dataMigrateInfo.MigrateFrom = migrateFrom
	dataMigrateInfo.MigrateTo = migrateTo

	j.Log.Info("dataMigrateInfo", "info", dataMigrateInfo)
	dataMigrateValue := cdatamigrate.DataMigrateValue{
		Name:            dataMigrate.Name,
		DataMigrateInfo: dataMigrateInfo,
	}
	dataMigrateValue.Owner = transfromer.GenerateOwnerReferenceFromObject(dataMigrate)

	// 6. create the value file
	data, err := yaml.Marshal(dataMigrateValue)
	if err != nil {
		return
	}
	j.Log.Info("dataMigrate value", "value", string(data))

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-migrate-values.yaml", dataMigrate.Namespace, dataMigrate.Name))
	if err != nil {
		return
	}
	err = os.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}

func addWorkerPodAntiAffinity(dataMigrateInfo *cdatamigrate.DataMigrateInfo, dataMigrate *datav1alpha1.DataMigrate) {
	releaseName := utils.GetDataMigrateReleaseName(dataMigrate.Name)
	appValue := fmt.Sprintf("%s-workers", releaseName)

	podAffinityTerm := corev1.WeightedPodAffinityTerm{
		Weight: 100,
		PodAffinityTerm: corev1.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appValue,
				},
			},
			TopologyKey: "kubernetes.io/hostname",
		},
	}

	// Affinity is nil
	if dataMigrateInfo.Affinity == nil {
		dataMigrateInfo.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					podAffinityTerm,
				},
			},
		}
		return
	}
	// Affinity not nil, PodAntiAffinity is nil
	if dataMigrateInfo.Affinity.PodAntiAffinity == nil {
		dataMigrateInfo.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				podAffinityTerm,
			},
		}
		return
	}
	// PodAntiAffinity not nil
	dataMigrateInfo.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
		append(dataMigrateInfo.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, podAffinityTerm)
}

func (j *JuiceFSEngine) setParallelMigrateOptions(dataMigrateInfo *cdatamigrate.DataMigrateInfo, dataMigrate *datav1alpha1.DataMigrate) error {
	var err error
	dataMigrateInfo.ParallelOptions = cdatamigrate.ParallelOptions{
		SSHPort:                cdatamigrate.DefaultSSHPort,
		SSHReadyTimeoutSeconds: cdatamigrate.DefaultSSHReadyTimeoutSeconds,
		SSHSecretName:          dataMigrate.Spec.ParallelOptions[cdatamigrate.SSHSecretName],
	}

	sshPort, exist := dataMigrate.Spec.ParallelOptions[cdatamigrate.SSHPort]
	if exist {
		dataMigrateInfo.ParallelOptions.SSHPort, err = strconv.Atoi(sshPort)
		if err != nil {
			j.Log.Error(err, "sshPort in the parallelOptions is not a int")
			return errors.Wrap(err, "sshPort in the parallelOptions is not a int")
		}
	}

	sshReadyTimeoutSeconds, exist := dataMigrate.Spec.ParallelOptions[cdatamigrate.SSHReadyTimeoutSeconds]
	if exist {
		dataMigrateInfo.ParallelOptions.SSHReadyTimeoutSeconds, err = strconv.Atoi(sshReadyTimeoutSeconds)
		if err != nil {
			j.Log.Error(err, "sshReadyTimeoutSeconds in the parallelOptions is not a int")
			return errors.Wrap(err, "sshReadyTimeoutSeconds in the parallelOptions is not a int")
		}
	}
	return nil
}

func (j *JuiceFSEngine) genDataUrl(data datav1alpha1.DataToMigrate, targetDataset *datav1alpha1.Dataset, info *cdatamigrate.DataMigrateInfo) (dataUrl string, err error) {
	if data.DataSet != nil {
		fsInfo, err := GetFSInfoFromConfigMap(j.Client, data.DataSet.Name, data.DataSet.Namespace)
		if err != nil {
			return "", err
		}
		info.Options[Edition] = fsInfo[Edition]
		if fsInfo[Edition] == CommunityEdition {
			info.EncryptOptions = append(info.EncryptOptions, datav1alpha1.EncryptOption{
				Name: "FLUID_METAURL",
				ValueFrom: datav1alpha1.EncryptOptionSource{
					SecretKeyRef: datav1alpha1.SecretKeySelector{
						Name: fsInfo[MetaurlSecret],
						Key:  fsInfo[MetaurlSecretKey],
					},
				},
			})
			u, err := url.Parse("jfs://FLUID_METAURL/")
			if err != nil {
				return "", err
			}
			u.Path = "/"
			mountPoint := targetDataset.Spec.Mounts[0].MountPoint
			subpath, err := ParseSubPathFromMountPoint(mountPoint)
			if err != nil {
				return "", err
			}
			if subpath != "/" {
				u.Path = subpath
			}
			if data.DataSet.Path != "" {
				u.Path = path.Join(u.Path, data.DataSet.Path)
				if strings.HasSuffix(data.DataSet.Path, "/") {
					u.Path = u.Path + "/"
				}
			}
			dataUrl = u.String()
		} else {
			u, err := url.Parse(fmt.Sprintf("jfs://%s/", fsInfo[Name]))
			if err != nil {
				return "", err
			}
			u.Path = "/"
			if data.DataSet.Path != "" {
				u.Path = data.DataSet.Path
			}
			dataUrl = u.String()
			if fsInfo[FormatCmd] != "" {
				info.Options[FormatCmd] = fsInfo[FormatCmd]
			}
			if fsInfo[TokenSecret] != "" && fsInfo[TokenSecretKey] != "" {
				info.EncryptOptions = append(info.EncryptOptions, datav1alpha1.EncryptOption{
					Name: "TOKEN",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: fsInfo[TokenSecret],
							Key:  fsInfo[TokenSecretKey],
						},
					},
				})
			}
			if fsInfo[AccessKeySecret] != "" && fsInfo[AccessKeySecretKey] != "" {
				info.EncryptOptions = append(info.EncryptOptions, datav1alpha1.EncryptOption{
					Name: "ACCESS_KEY",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: fsInfo[AccessKeySecret],
							Key:  fsInfo[AccessKeySecretKey],
						},
					},
				})
			}
			if fsInfo[SecretKeySecret] != "" && fsInfo[SecretKeySecretKey] != "" {
				info.EncryptOptions = append(info.EncryptOptions, datav1alpha1.EncryptOption{
					Name: "SECRET_KEY",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: fsInfo[SecretKeySecret],
							Key:  fsInfo[SecretKeySecretKey],
						},
					},
				})
			}
		}
		return dataUrl, nil
	}
	if data.ExternalStorage != nil {
		if common.IsFluidNativeScheme(data.ExternalStorage.URI) {

			if strings.HasPrefix(data.ExternalStorage.URI, common.VolumeScheme.String()) {
				var (
					volName = "native-vol"
					subPath = ""
					pvcName string
				)
				parts := strings.SplitN(strings.TrimPrefix(data.ExternalStorage.URI, common.VolumeScheme.String()), "/", 2)

				if len(parts) > 1 {
					// with subpath, e.g. pvc://my-pvc/path/to/dir
					pvcName = parts[0]
					subPath = parts[1]
				} else {
					// without subpath, e.g. pvc://my-pvc
					pvcName = parts[0]
				}
				info.NativeVolumes = append(info.NativeVolumes, corev1.Volume{
					Name: volName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				})
				info.NativeVolumeMounts = append(info.NativeVolumeMounts, corev1.VolumeMount{
					Name:      volName,
					SubPath:   subPath,
					MountPath: NativeVolumeMigratePath,
				})
				return NativeVolumeMigratePath, nil
			}
			// TODO: support host path scheme local://
		} else {
			u, err := url.Parse(data.ExternalStorage.URI)
			if err != nil {
				return "", err
			}
			var accessKey, secretKey, token string
			for _, encryptOption := range data.ExternalStorage.EncryptOptions {
				name := encryptOption.Name
				keyName := name
				switch name {
				case "access-key":
					accessKey = "${EXTERNAL_ACCESS_KEY}"
					keyName = "EXTERNAL_ACCESS_KEY"
				case "secret-key":
					secretKey = "${EXTERNAL_SECRET_KEY}"
					keyName = "EXTERNAL_SECRET_KEY"
				case "token":
					token = "${EXTERNAL_TOKEN}"
					keyName = "EXTERNAL_TOKEN"
				}
				info.EncryptOptions = append(info.EncryptOptions, datav1alpha1.EncryptOption{
					Name:      keyName,
					ValueFrom: encryptOption.ValueFrom,
				})
			}
			if token != "" {
				secretKey = fmt.Sprintf("%s:%s", secretKey, token)
			}
			u.User = url.UserPassword(accessKey, secretKey)
			decodedValue, _ := url.QueryUnescape(u.String())
			dataUrl = decodedValue
			j.Log.Info("dataUrl", "dataUrl", dataUrl)
			return dataUrl, nil
		}
	}
	return
}
