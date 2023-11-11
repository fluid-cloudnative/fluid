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
	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

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
