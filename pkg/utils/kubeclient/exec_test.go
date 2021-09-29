/*

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
	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"io"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/url"
	"os"
	"testing"
)

func TestInitClient(t *testing.T) {
	PathExistsTrue := func(path string) bool {
		return true
	}
	PathExistsFalse := func(path string) bool {
		return false
	}
	BuildConfigFromFlagsCommon := func(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
		return nil, nil
	}
	BuildConfigFromFlagsErr := func(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
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

	err := initClient()
	if err != nil {
		t.Error("fail to exec the initClient function")
	}
	err = os.Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")
	if err != nil {
		t.Errorf("expected no error, get %v", err)
	}

	err = gohook.Hook(utils.PathExists, PathExistsTrue, nil)
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

func TestExecWIthOptions(t *testing.T) {
	ExecuteCommon := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return nil
	}
	ExecuteErr := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return errors.New("fail to exec the function")
	}
	wrappedUnhookExecute := func() {
		err := gohook.UnHook(execute)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	options := ExecOptions{
		Command:       []string{"ll"},
		Namespace:     "default",
		PodName:       "hadoop",
		ContainerName: "hadoop",


		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	}

	err := gohook.Hook(execute, ExecuteErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecWithOptions(options)
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookExecute()

	err = gohook.Hook(execute, ExecuteCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecWithOptions(options)
	if err != nil {
		t.Errorf("expected nil, get %v", err)
	}
	wrappedUnhookExecute()
}

func TestExecCommandInContainerWithFullOutput(t *testing.T) {
	ExecuteCommon := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return nil
	}
	ExecuteErr := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return errors.New("fail to exec the function")
	}
	wrappedUnhookExecute := func() {
		err := gohook.UnHook(execute)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	options := ExecOptions{
		Command:       []string{"ll"},
		Namespace:     "default",
		PodName:       "hadoop",
		ContainerName: "hadoop",

		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	}

	err := gohook.Hook(execute, ExecuteErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecCommandInContainerWithFullOutput(options.PodName, options.ContainerName, options.Namespace, options.Command)
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookExecute()

	err = gohook.Hook(execute, ExecuteCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecCommandInContainerWithFullOutput(options.PodName, options.ContainerName, options.Namespace, options.Command)
	if err != nil {
		t.Errorf("expected nil, get %v", err)
	}
	wrappedUnhookExecute()
}

func TestExecShellInContainer(t *testing.T) {
	ExecuteCommon := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return nil
	}
	ExecuteErr := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return errors.New("fail to exec the function")
	}
	wrappedUnhookExecute := func() {
		err := gohook.UnHook(execute)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	options := ExecOptions{
		Namespace:     "default",
		PodName:       "hadoop",
		ContainerName: "hadoop",

		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	}

	err := gohook.Hook(execute, ExecuteErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecShellInContainer(options.PodName, options.ContainerName, options.Namespace, "ll")
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookExecute()

	err = gohook.Hook(execute, ExecuteCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecShellInContainer(options.PodName, options.ContainerName, options.Namespace, "ll")
	if err != nil {
		t.Errorf("expected nil, get %v", err)
	}
	wrappedUnhookExecute()
}

func TestExecCommandInContainer(t *testing.T) {
	ExecuteCommon := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return nil
	}
	ExecuteErr := func(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		return errors.New("fail to exec the function")
	}
	wrappedUnhookExecute := func() {
		err := gohook.UnHook(execute)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	options := ExecOptions{
		Command:       []string{"ll"},
		Namespace:     "default",
		PodName:       "hadoop",
		ContainerName: "hadoop",

		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	}

	err := gohook.Hook(execute, ExecuteErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecCommandInContainer(options.PodName, options.ContainerName, options.Namespace, options.Command)
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	wrappedUnhookExecute()

	err = gohook.Hook(execute, ExecuteCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = ExecCommandInContainer(options.PodName, options.ContainerName, options.Namespace, options.Command)
	if err != nil {
		t.Errorf("expected nil, get %v", err)
	}
	wrappedUnhookExecute()
}
