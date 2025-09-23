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
	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
)

// IgnoreAlreadyExists ignores already existes error
func IgnoreAlreadyExists(err error) error {
	if apierrs.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// IgnoreNotFound ignores not found
func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func IgnoreNoKindMatchError(err error) error {
	if apimeta.IsNoMatchError(err) {
		return nil
	}
	return err
}

// LoggingErrorExceptConflict logs error except for updating operation violates with etcd concurrency control
func LoggingErrorExceptConflict(logging logr.Logger, err error, info string, namespacedKey types.NamespacedName) (result error) {
	if apierrs.IsConflict(err) {
		log.Info("Retry later when update operation violates with apiserver concurrency control.",
			"error", err,
			"name", namespacedKey.Name,
			"namespace", namespacedKey.Namespace)
	} else {
		log.Error(err, info, "name", namespacedKey.Name,
			"namespace", namespacedKey.Namespace)
		result = err
	}
	return result
}
