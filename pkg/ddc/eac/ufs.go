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

package eac

import (
	"errors"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (e *EACEngine) UsedStorageBytes() (int64, error) {
	return 0, nil
}

func (e *EACEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (e *EACEngine) TotalStorageBytes() (int64, error) {
	response, err := e.describeDirQuota()
	if err != nil {
		return 0, err
	}
	return response.DirQuotaInfos[0].UserQuotaInfos[0].SizeReal * 1024 * 1024 * 1024, nil
}

func (e *EACEngine) TotalFileNums() (int64, error) {
	response, err := e.describeDirQuota()
	if err != nil {
		return 0, err
	}
	return response.DirQuotaInfos[0].UserQuotaInfos[0].FileCountReal, nil
}

func (e *EACEngine) ShouldCheckUFS() (should bool, err error) {
	accessKeyID, accessKeySecret, err := e.getEACSecret()
	if err != nil {
		return false, err
	}
	if len(accessKeyID) == 0 || len(accessKeySecret) == 0 {
		return false, nil
	}

	return true, nil
}

func (e *EACEngine) PrepareUFS() (err error) {
	_, err = e.setDirQuota()
	if err != nil {
		e.Log.Error(err, "Failed to set dir quota")
		return err
	}
	return
}

// ShouldUpdateUFS EACEngine hasn't support UpdateOnUFSChange
func (e *EACEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (e *EACEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}

func (e *EACEngine) getEACSecret() (accessKeyID string, accessKeySecret string, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}

	if len(dataset.Spec.Mounts) == 0 {
		err = errors.New("empty mount point")
		return
	}

	for _, encryptOption := range dataset.Spec.Mounts[0].EncryptOptions {
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
		if err != nil {
			e.Log.Info("can't get the secret",
				"namespace", e.namespace,
				"name", e.name,
				"secretName", secretKeyRef.Name)
			return "", "", err
		}

		value, ok := secret.StringData[secretKeyRef.Key]
		if !ok {
			err = fmt.Errorf("can't get %s from secret %s namespace %s", secretKeyRef.Key, secretKeyRef.Name, e.namespace)
			return "", "", err
		}

		switch encryptOption.Name {
		case AccessKeyIDName:
			accessKeyID = value
			break
		case AccessKeySecretName:
			accessKeySecret = value
			break
		default:
			break
		}
	}

	return
}
