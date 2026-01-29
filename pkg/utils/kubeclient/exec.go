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

package kubeclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	securityutils "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// https://github.com/kubernetes/kubernetes/blob/v1.6.1/test/e2e/framework/exec_util.go
// Global variables
var (
	clientset      *kubernetes.Clientset
	restConfig     *restclient.Config
	log            logr.Logger = ctrl.Log.WithName("kubeclient")
	kubeconfigPath             = "~/.kube/config"
	mutex                      = &sync.Mutex{}
)

// ExecOptions passed to ExecWithOptions
type ExecOptions struct {
	Command []string

	Namespace     string
	PodName       string
	ContainerName string

	Stdin         io.Reader
	CaptureStdout bool
	CaptureStderr bool
	// If false, whitespace in std{err,out} will be removed.
	PreserveWhitespace bool
}

func (opt ExecOptions) String() string {
	return fmt.Sprintf(
		"{Command: %v, Namespace: %v, PodName: %v, ContainerName: %v, CaptureStdout: %v, CaptureStderr: %v, PreserveWhitespace: %v}",
		securityutils.FilterCommand(opt.Command),
		opt.Namespace,
		opt.PodName,
		opt.ContainerName,
		opt.CaptureStdout,
		opt.CaptureStderr,
		opt.PreserveWhitespace,
	)
}

func initClient() error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	if restConfig == nil {

		home, err := utils.Home()
		if err != nil {
			return err
		}
		kubeconfigPath = path.Join(home, ".kube/config")
		if len(os.Getenv(common.RecommendedKubeConfigPathEnv)) > 0 {
			kubeconfigPath = os.Getenv(common.RecommendedKubeConfigPathEnv)
		}
		if !utils.PathExists(kubeconfigPath) {
			kubeconfigPath = ""
		}
		log.Info("kubeconfig file is placed.", "config", kubeconfigPath)
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return err
		}
	}
	if clientset == nil {
		clientset, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecWithOptions executes a command in the specified container,
// returning stdout, stderr and error. `options` allowed for
// additional parameters to be passed.
func ExecWithOptions(ctx context.Context, options ExecOptions) (string, string, error) {
	err := cmdguard.ValidateCommandSlice(options.Command)
	if err != nil {
		return "", "", err
	}

	err = initClient()
	if err != nil {
		return "", "", err
	}

	log.V(1).Info("ExecWithOptions", "ExecWithOptions", options)

	const tty = false

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.ContainerName)
	req.VersionedParams(&v1.PodExecOptions{
		Container: options.ContainerName,
		Command:   options.Command,
		Stdin:     options.Stdin != nil,
		Stdout:    options.CaptureStdout,
		Stderr:    options.CaptureStderr,
		TTY:       tty,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err = doExecute(ctx, "POST", req.URL(), restConfig, options.Stdin, &stdout, &stderr, tty)

	if options.PreserveWhitespace {
		return stdout.String(), stderr.String(), err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// ExecCommandInContainerWithFullOutput executes a command in the
// specified container and return stdout, stderr and error
func ExecCommandInContainerWithFullOutput(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, err error) {
	return ExecWithOptions(ctx, ExecOptions{
		Command:       cmd,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,

		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	})
}

// Exec commands in container without any timeout.
func ExecCommandInContainer(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, err error) {
	return ExecCommandInContainerWithFullOutput(context.Background(), podName, containerName, namespace, cmd)
}

// execResult encapsulates the result of a container exec operation.
// This struct is used to safely pass results through a channel,
// avoiding data races on shared variables.
type execResult struct {
	stdout string
	stderr string
	err    error
}

// Exec commands in container with a given timeout.
// This function is thread-safe and avoids data races by using a result channel
// instead of writing to shared return variables from a goroutine.
func ExecCommandInContainerWithTimeout(podName string, containerName string, namespace string, cmd []string, timeout time.Duration) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	// Use a buffered channel to prevent goroutine leak when timeout occurs.
	// The goroutine can always send its result even if no one is receiving.
	resultCh := make(chan execResult, 1)

	go func() {
		out, errOut, execErr := ExecCommandInContainerWithFullOutput(ctx, podName, containerName, namespace, cmd)
		resultCh <- execResult{stdout: out, stderr: errOut, err: execErr}
	}()

	select {
	case result := <-resultCh:
		// Command completed within timeout
		return result.stdout, result.stderr, result.err
	case <-ctx.Done():
		// Timeout occurred - return timeout error immediately.
		// The goroutine will eventually complete (context cancellation will propagate)
		// and send to the buffered channel, which will be garbage collected.
		return "", "", fmt.Errorf("timed out for %v", timeout)
	}
}

func doExecute(ctx context.Context, method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.StreamWithContext(ctx,
		remotecommand.StreamOptions{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
			Tty:    tty,
		})
}
