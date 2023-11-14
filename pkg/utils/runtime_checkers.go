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

package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CheckObject(client client.Client, key types.NamespacedName, obj client.Object) (found bool, err error) {
	if err = client.Get(context.TODO(), key, obj); err != nil {
		if IgnoreNoKindMatchError(err) == nil || IgnoreNotFound(err) == nil {
			log.V(1).Info("Object not exists yet, skip.", "runtime", obj)
			err = nil
		}
		return
	}
	found = true
	log.Info("Succeed in finding the object", "runtime", obj)
	return
}
