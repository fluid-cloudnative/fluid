/*
Copyright 2020 The Fluid Authors.

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

package common

type datasetMountsLengthConstraint struct {
	AllowZeroMountPoints      bool
	AllowUnlimitedMountPoints bool
	AllowedMaxMountsLength    int
}

var MountsLengthConstraintMap = map[string]datasetMountsLengthConstraint{
	AlluxioEngineImpl:    {false, true, -1},
	JindoFSEngineImpl:    {false, true, -1},
	JindoFSxEngineImpl:   {false, true, -1},
	JindoCacheEngineImpl: {false, true, -1},
	GooseFSEngineImpl:    {false, true, -1},
	JuiceFSEngineImpl:    {false, true, -1},
	ThinEngineImpl:       {true, true, -1},
	EFCEngineImpl:        {false, true, -1},
	VineyardEngineImpl:   {true, true, -1},
}
