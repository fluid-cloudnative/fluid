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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestLoggingErrorExceptConflict(t *testing.T) {
	logger := fake.NullLogger()
	result := LoggingErrorExceptConflict(logger,
		apierrors.NewConflict(schema.GroupResource{},
			"test",
			fmt.Errorf("the object has been modified; please apply your changes to the latest version and try again")),
		"Failed to setup worker",
		types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result != nil {
		t.Errorf("Expected error result is null, but got %v", result)
	}
	result = LoggingErrorExceptConflict(logger,
		apierrors.NewNotFound(schema.GroupResource{}, "test"),
		"Failed to setup worker", types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result == nil {
		t.Errorf("Expected error result is not null, but got %v", result)
	}
}
