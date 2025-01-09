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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"hash/fnv"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"strings"
)

func ComputeHash(object interface{}) (string, error) {
	objString, err := json.Marshal(object)
	if err != nil {
		return "", errors.Wrap(err, "failed to compute hash.")
	}

	// hasher := sha1.New()
	hasher := fnv.New32()
	if _, err := io.Copy(hasher, bytes.NewReader(objString)); err != nil {
		return "", errors.Wrapf(err, "failed to compute hash for sha256. [%s]", objString)
	}

	sha := hasher.Sum32()
	return rand.SafeEncodeString(fmt.Sprint(sha)), nil
}

func ComputeFullNamespacedNameHashValue(namespace, name string) (string, error) {
	return ComputeHash(v1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
	})
}

// TransferFullNamespacedNameWithPrefixToLegalValue Transfer a fully namespaced name with a prefix to a legal value which under max length limit.
// If the full namespaced name exceeds 63 characters, it calculates the hash value of the name and truncates the name and namespace,
// then appends the hash value to ensure the name's uniqueness and length constraint.
func TransferFullNamespacedNameWithPrefixToLegalValue(prefix, namespace, name string) (fullNamespacedName string) {
	fullNamespacedName = fmt.Sprintf("%s%s-%s", prefix, namespace, name)

	// ensure forward compatibility
	if len(fullNamespacedName) < 63 {
		return
	}

	namespacedNameHashValue, err := ComputeFullNamespacedNameHashValue(namespace, name)
	if err != nil {
		log.Error(err, "fail to compute hash value for namespacedName, and fall back to the original value which will cause the failure of resource creation.")
		return
	}
	trimMetadata := func(s string) string {
		s = strings.ReplaceAll(s, "-", "")
		if len(s) <= 8 {
			return s
		}
		return s[:8]
	}
	fullNamespacedName = fmt.Sprintf("%s%s-%s-%s", prefix, trimMetadata(namespace), trimMetadata(name), namespacedNameHashValue)

	return
}
