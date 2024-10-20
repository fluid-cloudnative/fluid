/*
Copyright 2023 The Fluid Author.

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
