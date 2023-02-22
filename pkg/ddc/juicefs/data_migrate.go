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
	"net/url"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
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
		valueFileName, err := j.generateDataMigrateValueFile(ctx, targetDataMigrate)
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
		ctx.Recorder.Eventf(&targetDataMigrate, corev1.EventTypeNormal, common.DataLoadJobStarted, "The DataMigrate job %s started", jobName)
	}
	return err
}

func (j *JuiceFSEngine) generateDataMigrateValueFile(r cruntime.ReconcileRequestContext, dataMigrate datav1alpha1.DataMigrate) (valueFileName string, err error) {
	// 1. get the target dataset
	targetDataset, err := utils.GetTargetDatasetOfMigrate(r.Client, dataMigrate)
	if err != nil {
		return "", err
	}
	j.Log.Info("target dataset", "dataset", targetDataset)

	// 2. get info
	imageName, imageTag := dataMigrate.Spec.Image, dataMigrate.Spec.ImageTag

	if len(imageName) == 0 {
		defaultImageInfo := strings.Split(common.DefaultJuiceFSMigrateImage, ":")
		if len(defaultImageInfo) < 1 {
			panic("invalid default dataload image!")
		} else {
			imageName = defaultImageInfo[0]
		}
	}

	if len(imageTag) == 0 {
		defaultImageInfo := strings.Split(common.DefaultJuiceFSRuntimeImage, ":")
		if len(defaultImageInfo) < 2 {
			panic("invalid default dataload image!")
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
	migrateFrom, err := j.genDataUrl(dataMigrate.Spec.From, &dataMigrateInfo)
	if err != nil {
		return "", err
	}
	migrateTo, err := j.genDataUrl(dataMigrate.Spec.To, &dataMigrateInfo)
	if err != nil {
		return "", err
	}
	dataMigrateInfo.MigrateFrom = migrateFrom
	dataMigrateInfo.MigrateTo = migrateTo

	j.Log.Info("dataMigrateInfo", "info", dataMigrateInfo)
	dataMigrateValue := cdatamigrate.DataMigrateValue{
		DataMigrateInfo: dataMigrateInfo,
	}

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

func (j *JuiceFSEngine) genDataUrl(data datav1alpha1.DataToMigrate, info *cdatamigrate.DataMigrateInfo) (dataUrl string, err error) {
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
			if data.DataSet.Path != "" {
				u.Path = data.DataSet.Path
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
		}
		return dataUrl, nil
	}
	if data.ExternalStorage != nil {
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
				accessKey = "${ACCESS_KEY}"
				keyName = "ACCESS_KEY"
			case "secret-key":
				secretKey = "${SECRET_KEY}"
				keyName = "SECRET_KEY"
			case "token":
				token = "${TOKEN}"
				keyName = "TOKEN"
			}
			secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
			_, err := kubeclient.GetSecret(j.Client, secretKeyRef.Name, j.namespace)
			if err != nil {
				j.Log.Info("can't get the secret",
					"namespace", j.namespace,
					"name", j.name,
					"secretName", secretKeyRef.Name)
				return "", err
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
	return
}
