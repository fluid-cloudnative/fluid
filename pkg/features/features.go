/*
Copyright 2026 The Fluid Authors.

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

package features

import (
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

const (
	// GracefulWorkerScaleDown gates graceful worker scale-down for AlluxioRuntime
	// on a standard Kubernetes StatefulSet.
	//
	// When enabled, workers targeted for removal are decommissioned from the
	// Alluxio cluster before the pod is terminated, giving the master time to
	// migrate their cached blocks to the surviving workers. Without this gate,
	// cached data held on removed workers is lost immediately on scale-in.
	//
	// This only supports the standard StatefulSet scale-down order (highest
	// ordinal first); it does not yet integrate with OpenKruise's Advanced
	// StatefulSet for selective deletion of non-highest-ordinal pods.
	GracefulWorkerScaleDown featuregate.Feature = "GracefulWorkerScaleDown"
)

var defaultFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	GracefulWorkerScaleDown: {Default: false, PreRelease: featuregate.Alpha},
}

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultFeatureGates))
}
