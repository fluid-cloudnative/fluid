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

package kubeclient

import (
	"errors"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestInitClient(t *testing.T) {
	PathExistsTrue := func(path string) bool {
		return true
	}
	PathExistsFalse := func(path string) bool {
		return false
	}
	BuildConfigFromFlagsCommon := func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
		return nil, nil
	}
	BuildConfigFromFlagsErr := func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
		return nil, errors.New("fail to run the function")
	}
	NewForConfigCommon := func(c *rest.Config) (*kubernetes.Clientset, error) {
		return nil, nil
	}
	NewForConfigError := func(c *rest.Config) (*kubernetes.Clientset, error) {
		return nil, errors.New("fail to run the function")
	}
	wrappedUnhookPathExists := func() {
		err := gohook.UnHook(utils.PathExists)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookBuildConfigFromFlags := func() {
		err := gohook.UnHook(clientcmd.BuildConfigFromFlags)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookNewForConfig := func() {
		err := gohook.UnHook(kubernetes.NewForConfig)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	t.Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")

	err := gohook.Hook(utils.PathExists, PathExistsTrue, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookBuildConfigFromFlags()

	err = gohook.Hook(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubernetes.NewForConfig, NewForConfigError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookNewForConfig()

	err = gohook.Hook(kubernetes.NewForConfig, NewForConfigCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	restConfig = nil
	clientset = nil

	err = initClient()
	if err != nil {
		t.Errorf("expected no error, get %v", err)
	}
	wrappedUnhookNewForConfig()
	wrappedUnhookBuildConfigFromFlags()
	wrappedUnhookPathExists()

	err = gohook.Hook(utils.PathExists, PathExistsFalse, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookBuildConfigFromFlags()

	err = gohook.Hook(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubernetes.NewForConfig, NewForConfigError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookNewForConfig()

	err = gohook.Hook(kubernetes.NewForConfig, NewForConfigCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	restConfig = nil
	clientset = nil

	err = initClient()
	if err != nil {
		t.Errorf("expected no error, get %v", err)
	}
	wrappedUnhookNewForConfig()
	wrappedUnhookBuildConfigFromFlags()
	wrappedUnhookPathExists()

}
