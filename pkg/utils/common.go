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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/pkg/errors"
)

func GetEnvByKey(k string) (string, error) {
	if v, ok := os.LookupEnv(k); ok {
		return v, nil
	} else {
		return "", errors.Errorf("can not find the env value, key:%s", k)
	}
}

// determine if subPath is a subdirectory of path
func IsSubPath(path, subPath string) bool {
	rel, err := filepath.Rel(path, subPath)

	if err != nil {
		return false
	}

	if strings.HasPrefix(rel, "..") {
		return false
	}

	return true
}

func GetObjectMeta(object client.Object) (objectMeta metav1.Object, err error) {
	objectMetaAccessor, isOM := object.(metav1.ObjectMetaAccessor)
	if !isOM {
		err = fmt.Errorf("object is not ObjectMetaAccessor")
		return
	}
	objectMeta = objectMetaAccessor.GetObjectMeta()
	return
}
