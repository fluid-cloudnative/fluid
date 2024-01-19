/*

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

	// cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	cmd, err := utils.PipeCommand("bash", "-c", strings.Join(args, " "))
	if err != nil {
		return "", err
	}
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
