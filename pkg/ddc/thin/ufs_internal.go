/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package thin

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
)

func (t *ThinEngine) totalStorageBytesInternal() (total int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	total, err = fileUtils.GetUsedSpace(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) totalFileNumsInternal() (fileCount int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	fileCount, err = fileUtils.GetFileCount(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) usedSpaceInternal() (usedSpace int64, err error) {
	stsName := t.getFuseDaemonsetName()
	pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewThinFileUtils(pods[0].Name, common.ThinFuseContainer, t.namespace, t.Log)
	usedSpace, err = fileUtils.GetUsedSpace(t.getTargetPath())
	if err != nil {
		return
	}

	return
}

// genFuseMountOptions extracts mount options needed by Thin FUSE from mount info, including options and encrypt options. If extractEncryptOptions is set to false,
// encrypt options will be skipped and won't be extracted
func (t *ThinEngine) genFuseMountOptions(m datav1alpha1.Mount, SharedOptions map[string]string, SharedEncryptOptions []datav1alpha1.EncryptOption, extractEncryptOptions bool) (map[string]string, error) {

	// initialize mount options
	mOptions := map[string]string{}
	if len(SharedOptions) > 0 {
		mOptions = SharedOptions
	}
	for key, value := range m.Options {
		mOptions[key] = value
	}

	// if encryptOptions have the same key with options
	// it will overwrite the corresponding value
	if extractEncryptOptions {
		var err error
		mOptions, err = t.genEncryptOptions(SharedEncryptOptions, mOptions, m.Name)
		if err != nil {
			return mOptions, err
		}

		//gen public encryptOptions
		mOptions, err = t.genEncryptOptions(m.EncryptOptions, mOptions, m.Name)
		if err != nil {
			return mOptions, err
		}
	}

	return mOptions, nil
}

// thin encrypt mount options
func (t *ThinEngine) genEncryptOptions(EncryptOptions []datav1alpha1.EncryptOption, mOptions map[string]string, name string) (map[string]string, error) {
	for _, item := range EncryptOptions {

		if _, ok := mOptions[item.Name]; ok {
			err := fmt.Errorf("the option %s is set more than one times, please double check the dataset's option and encryptOptions", item.Name)
			return mOptions, err
		}

		securityutil.UpdateSensitiveKey(item.Name)
		sRef := item.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(t.Client, sRef.Name, t.namespace)
		if err != nil {
			t.Log.Error(err, "get secret by mount encrypt options failed", "name", item.Name)
			return mOptions, err
		}

		t.Log.Info("get value from secret", "mount name", name, "secret key", sRef.Key)

		v := secret.Data[sRef.Key]
		mOptions[item.Name] = string(v)
	}

	return mOptions, nil
}
