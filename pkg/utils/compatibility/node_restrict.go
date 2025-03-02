/*
Copyright 2025 The Fluid Authors.

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

package compatibility

import (
	"github.com/blang/semver/v4"

	nativeLog "log"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	nodeBindingTokenSupported = false
	nodeBindingTokenOnce      sync.Once
)

// Beta release, default enabled. see https://github.com/kubernetes/enhancements/issues/4193
const nodeBindingTokenSupportedVersion = "v1.30.0"

// Checks the ServiceAccountTokenPodNodeInfo feature gate, whether the apiserver embeds the node name for the associated node when issuing service account tokens bound to Pod objects.
func discoverNodeBindingTokenCompatibility() {
	nativeLog.Printf("Discovering k8s version to check NodeBindingToken compatibility...")
	restConfig := ctrl.GetConfigOrDie()
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(restConfig)

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil && !errors.IsNotFound(err) {
		nativeLog.Fatalf("failed to discover batch/v1 group version: %v", err)
	}
	// transform to semver.Version and compare
	currentVersion, err := semver.ParseTolerant(serverVersion.GitVersion)
	if err != nil {
		nativeLog.Fatalf("Failed to parse current version: %v", err)
	}
	targetVersion, err := semver.ParseTolerant(nodeBindingTokenSupportedVersion)
	if err != nil {
		nativeLog.Fatalf("Failed to parse target version: %v", err)
	}

	if currentVersion.GTE(targetVersion) {
		nodeBindingTokenSupported = true
	}
}

func IsNodeBindingTokenSupported() bool {
	nodeBindingTokenOnce.Do(func() {
		discoverNodeBindingTokenCompatibility()
	})
	return nodeBindingTokenSupported
}
