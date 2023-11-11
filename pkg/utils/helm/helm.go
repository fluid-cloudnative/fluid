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

package helm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("helm")
}

var helmCmd = []string{"ddc-helm"}

// GenerateValueFile generates value file.
// It returns the name of the value file and error
func GenerateValueFile(values interface{}) (valueFileName string, err error) {
	// 1. generate the template file
	valueFile, err := os.CreateTemp(os.TempDir(), "values")
	if err != nil {
		log.Error(err, "Failed to create tmp file", "tmpfile", valueFile.Name())
		return "", err
	}

	valueFileName = valueFile.Name()
	log.V(1).Info("Save the values file", "fileName", valueFileName)

	// 2. dump the object into the template file
	err = utils.ToYaml(values, valueFile)
	return valueFileName, err
}

// GenerateHelmTemplate generates helm template without tiller: helm template -f values.yaml chart_name
// Exec /usr/local/bin/helm, [template -f /tmp/values313606961 --namespace default --name hj /charts/tf-horovod]
// returns generated template file: templateFileName
func GenerateHelmTemplate(name string, namespace string, valueFileName string, chartName string, options ...string) (templateFileName string, err error) {
	tempName := fmt.Sprintf("%s.yaml", name)
	templateFile, err := os.CreateTemp("", tempName)
	if err != nil {
		return templateFileName, err
	}
	templateFileName = templateFile.Name()

	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return templateFileName, err
	}

	// 3. check if the chart file exists
	// if strings.HasPrefix(chartName, "/") {
	if _, err = os.Stat(chartName); err != nil {
		return templateFileName, err
	}
	// }

	// 4. prepare the arguments
	args := []string{binary, "template", "-f", valueFileName,
		"--namespace", namespace,
		"--name", name,
	}
	if len(options) != 0 {
		args = append(args, options...)
	}

	args = append(args, []string{chartName, ">", templateFileName}...)

	log.V(1).Info("Exec bash -c ", "cmd", args)

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	log.V(1).Info("Generating template", "args", args)
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	// cmd.Env = env
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", string(out))
	if err != nil {
		return templateFileName, fmt.Errorf("failed to execute %s, %v with %v", binary, args, err)
	}

	// // 6. clean up the value file if needed
	// if log.GetLevel() != log.DebugLevel {
	// 	err = os.Remove(valueFileName)
	// 	if err != nil {
	// 		log.Warnf("Failed to delete %s due to %v", valueFileName, err)
	// 	}
	// }

	return templateFileName, nil
}

// GetChartVersion checks the chart version by given the chart directory
// helm inspect chart /charts/tf-horovod
func GetChartVersion(chart string) (version string, err error) {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return "", err
	}

	// 1. check if the chart file exists, if it's it's unix path, then check if it's exist
	// if strings.HasPrefix(chart, "/") {
	if _, err = os.Stat(chart); err != nil {
		return "", err
	}
	// }

	// 2. prepare the arguments
	args := []string{binary, "inspect", "chart", chart,
		"|", "grep", "version:"}
	log.V(1).Info("Exec bash -c", "args", args)

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("failed to find version when executing %s, result is %s", args, out)
	}
	fields := strings.Split(lines[0], ":")
	if len(fields) != 2 {
		return "", fmt.Errorf("failed to find version when executing %s, result is %s", args, out)
	}

	version = strings.TrimSpace(fields[1])
	return version, nil
}

// GetChartName extracts the last element of the chart's path as the chart's name
func GetChartName(chart string) string {
	return filepath.Base(chart)
}
