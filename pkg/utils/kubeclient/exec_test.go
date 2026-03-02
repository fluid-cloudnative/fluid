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
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
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

	t.Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")

	pathExistPatch := gomonkey.ApplyFunc(utils.PathExists, PathExistsTrue)
	buildConfigFromFlagsPatch := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsErr)

	restConfig = nil
	clientset = nil

	err := initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	buildConfigFromFlagsPatch.Reset()

	buildConfigFromFlagsPatch.ApplyFunc(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsCommon)
	newForConfigPatch := gomonkey.ApplyFunc(kubernetes.NewForConfig, NewForConfigError)
	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	newForConfigPatch.Reset()

	newForConfigPatch.ApplyFunc(kubernetes.NewForConfig, NewForConfigCommon)

	restConfig = nil
	clientset = nil

	err = initClient()
	if err != nil {
		t.Errorf("expected no error, get %v", err)
	}
	newForConfigPatch.Reset()
	buildConfigFromFlagsPatch.Reset()
	pathExistPatch.Reset()

	pathExistPatch.ApplyFunc(utils.PathExists, PathExistsFalse)
	buildConfigFromFlagsPatch.ApplyFunc(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsErr)

	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	buildConfigFromFlagsPatch.Reset()

	buildConfigFromFlagsPatch.ApplyFunc(clientcmd.BuildConfigFromFlags, BuildConfigFromFlagsCommon)
	newForConfigPatch.ApplyFunc(kubernetes.NewForConfig, NewForConfigError)
	restConfig = nil
	clientset = nil

	err = initClient()
	if err == nil {
		t.Errorf("expected error, get nil")
	}
	newForConfigPatch.Reset()

	newForConfigPatch.ApplyFunc(kubernetes.NewForConfig, NewForConfigCommon)
	restConfig = nil
	clientset = nil

	err = initClient()
	if err != nil {
		t.Errorf("expected no error, get %v", err)
	}
	newForConfigPatch.Reset()
	buildConfigFromFlagsPatch.Reset()
	pathExistPatch.Reset()
}

// TestExecCommandInContainerWithTimeout_Success tests that the function returns
// correctly when the command completes before the timeout.
func TestExecCommandInContainerWithTimeout_Success(t *testing.T) {
	expectedStdout := "test output"
	expectedStderr := "test error output"

	// Mock ExecCommandInContainerWithFullOutput to return immediately with success
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			return expectedStdout, expectedStderr, nil
		})
	defer patch.Reset()

	stdout, stderr, err := ExecCommandInContainerWithTimeout(
		"test-pod", "test-container", "test-namespace",
		[]string{"echo", "hello"}, 5*time.Second)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stdout != expectedStdout {
		t.Errorf("expected stdout %q, got %q", expectedStdout, stdout)
	}
	if stderr != expectedStderr {
		t.Errorf("expected stderr %q, got %q", expectedStderr, stderr)
	}
}

// TestExecCommandInContainerWithTimeout_Timeout tests that the function returns
// a timeout error when the command takes longer than the specified timeout.
func TestExecCommandInContainerWithTimeout_Timeout(t *testing.T) {
	// Mock ExecCommandInContainerWithFullOutput to block until context is cancelled
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			// Wait for context cancellation (simulating a slow command)
			<-ctx.Done()
			return "should not see this", "should not see this either", ctx.Err()
		})
	defer patch.Reset()

	timeout := 100 * time.Millisecond
	start := time.Now()

	stdout, stderr, err := ExecCommandInContainerWithTimeout(
		"test-pod", "test-container", "test-namespace",
		[]string{"sleep", "10"}, timeout)

	elapsed := time.Since(start)

	// Verify timeout occurred
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got %v", err)
	}

	// Verify stdout and stderr are empty on timeout
	if stdout != "" {
		t.Errorf("expected empty stdout on timeout, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr on timeout, got %q", stderr)
	}

	// Verify the function returned promptly after timeout
	if elapsed > 2*timeout {
		t.Errorf("function took too long to return after timeout: %v", elapsed)
	}
}

// TestExecCommandInContainerWithTimeout_ErrorPropagation tests that errors from
// the underlying exec function are properly propagated.
func TestExecCommandInContainerWithTimeout_ErrorPropagation(t *testing.T) {
	expectedErr := errors.New("exec failed: command not found")

	// Mock ExecCommandInContainerWithFullOutput to return an error
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			return "", "command not found", expectedErr
		})
	defer patch.Reset()

	_, stderr, err := ExecCommandInContainerWithTimeout(
		"test-pod", "test-container", "test-namespace",
		[]string{"nonexistent-command"}, 5*time.Second)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr.Error(), err.Error())
	}
	if stderr != "command not found" {
		t.Errorf("expected stderr 'command not found', got %q", stderr)
	}
}

// TestExecCommandInContainerWithTimeout_NoDataRace tests that concurrent calls
// to the function don't cause data races. This test should be run with -race flag.
func TestExecCommandInContainerWithTimeout_NoDataRace(t *testing.T) {
	var callCount int32

	// Mock ExecCommandInContainerWithFullOutput with varying delays
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			count := atomic.AddInt32(&callCount, 1)
			// Vary the execution time to increase chance of race detection
			delay := time.Duration(count%3) * 10 * time.Millisecond
			select {
			case <-time.After(delay):
				return "stdout-" + podName, "stderr-" + podName, nil
			case <-ctx.Done():
				return "", "", ctx.Err()
			}
		})
	defer patch.Reset()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple concurrent calls
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			podName := "pod-" + string(rune('a'+id%26))
			stdout, stderr, err := ExecCommandInContainerWithTimeout(
				podName, "container", "namespace",
				[]string{"test"}, 500*time.Millisecond)

			// Verify results are consistent (not corrupted by races)
			if err == nil {
				if !strings.HasPrefix(stdout, "stdout-") {
					t.Errorf("goroutine %d: unexpected stdout prefix: %q", id, stdout)
				}
				if !strings.HasPrefix(stderr, "stderr-") {
					t.Errorf("goroutine %d: unexpected stderr prefix: %q", id, stderr)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestExecCommandInContainerWithTimeout_GoroutineLeak tests that goroutines
// don't leak when timeout occurs before the command completes.
func TestExecCommandInContainerWithTimeout_GoroutineLeak(t *testing.T) {
	var activeGoroutines int32

	// Mock ExecCommandInContainerWithFullOutput to track active goroutines
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			atomic.AddInt32(&activeGoroutines, 1)
			defer atomic.AddInt32(&activeGoroutines, -1)

			// Simulate a slow operation that respects context cancellation
			select {
			case <-time.After(1 * time.Second):
				return "completed", "", nil
			case <-ctx.Done():
				return "", "", ctx.Err()
			}
		})
	defer patch.Reset()

	// Launch several calls that will timeout
	const numCalls = 10
	var wg sync.WaitGroup
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		go func() {
			defer wg.Done()
			_, _, _ = ExecCommandInContainerWithTimeout(
				"pod", "container", "namespace",
				[]string{"slow-command"}, 50*time.Millisecond)
		}()
	}

	wg.Wait()

	// Wait a bit for goroutines to clean up after context cancellation
	time.Sleep(200 * time.Millisecond)

	// Verify no goroutines are leaked
	remaining := atomic.LoadInt32(&activeGoroutines)
	if remaining != 0 {
		t.Errorf("expected 0 active goroutines after cleanup, got %d", remaining)
	}
}

// TestExecCommandInContainerWithTimeout_ContextCancellation tests that context
// cancellation is properly propagated to the underlying exec function.
func TestExecCommandInContainerWithTimeout_ContextCancellation(t *testing.T) {
	var ctxCancelled bool
	var mu sync.Mutex

	// Mock ExecCommandInContainerWithFullOutput to check context cancellation
	patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
		func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
			// Wait for context to be cancelled
			<-ctx.Done()
			mu.Lock()
			ctxCancelled = true
			mu.Unlock()
			return "", "", ctx.Err()
		})
	defer patch.Reset()

	// Use a very short timeout
	_, _, err := ExecCommandInContainerWithTimeout(
		"pod", "container", "namespace",
		[]string{"command"}, 10*time.Millisecond)

	// Give goroutine time to observe cancellation
	time.Sleep(50 * time.Millisecond)

	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", err)
	}

	mu.Lock()
	if !ctxCancelled {
		t.Error("expected context to be cancelled, but it wasn't")
	}
	mu.Unlock()
}

// TestExecResult tests the execResult struct.
func TestExecResult(t *testing.T) {
	result := execResult{
		stdout: "test stdout",
		stderr: "test stderr",
		err:    errors.New("test error"),
	}

	if result.stdout != "test stdout" {
		t.Errorf("expected stdout 'test stdout', got %q", result.stdout)
	}
	if result.stderr != "test stderr" {
		t.Errorf("expected stderr 'test stderr', got %q", result.stderr)
	}
	if result.err == nil || result.err.Error() != "test error" {
		t.Errorf("expected error 'test error', got %v", result.err)
	}
}
