/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
