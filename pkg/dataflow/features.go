/*
Copyright 2024 The Fluid Authors.

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

package dataflow

import (
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

const (
	// DataflowAffinity Enable affinity inheritance for dataflow operations.
	DataflowAffinity featuregate.Feature = "DataflowAffinity"
)

// defaultFeatureGates consists of all known fluid-specific feature keys.
// To add a new feature, define a key for it above and add it here.
var defaultFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	DataflowAffinity: {Default: false, PreRelease: featuregate.Beta},
}

func init() {
	// register the feature gates and its default value.
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultFeatureGates))
}

// Enabled is helper for `utilfeature.DefaultFeatureGate.Enabled()`
func Enabled(f featuregate.Feature) bool {
	return utilfeature.DefaultFeatureGate.Enabled(f)
}
