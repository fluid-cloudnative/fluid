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
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDataProcess(client client.Client, name, namespace string) (*datav1alpha1.DataProcess, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	var dataprocess datav1alpha1.DataProcess
	if err := client.Get(context.TODO(), key, &dataprocess); err != nil {
		return nil, err
	}

	return &dataprocess, nil
}

// GetDataProcessReleaseName returns the helm release name given the DataProcess's name.
func GetDataProcessReleaseName(name string) string {
	return fmt.Sprintf("%s-processor", name)
}

func GetDataProcessJobName(releaseName string) string {
	return fmt.Sprintf("%s-job", releaseName)
}
