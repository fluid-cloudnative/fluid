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
	"os"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	yaml "gopkg.in/yaml.v2"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("utils")
}

// ToYaml converts values from json format to yaml format and stores the values to the file.
// It will return err when failed to marshal value or write file.
func ToYaml(values interface{}, file *os.File) error {
	log.V(1).Info("create yaml file", "values", values)
	data, err := yaml.Marshal(values)
	if err != nil {
		log.Error(err, "failed to marshal value", "value", values)
		return err
	}

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		log.Error(err, "failed to write file", "data", data, "fileName", file.Name())
	}
	return err
}
