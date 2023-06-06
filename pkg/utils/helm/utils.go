/*
Copyright 2023 The Fluid Author.

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
	"time"

	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
)

// InstallRelease installs the release with cmd: helm install -f values.yaml chart_name, support helm v3
func InstallRelease(name string, namespace string, valueFile string, chartName string) error {
	defer utils.TimeTrack(time.Now(), "Helm.InstallRelease", "name", name, "namespace", namespace)
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	// 3. check if the chart file exists, if it's it's unix path, then check if it's exist
	if strings.HasPrefix(chartName, "/") {
		if _, err = os.Stat(chartName); os.IsNotExist(err) {
			// TODO: the chart will be put inside the binary in future
			return err
		}
	}

	// 4. prepare the arguments
	args := []string{"install", "-f", valueFile, "--namespace", namespace, name, chartName}

	// env := os.Environ()
	// if types.KubeConfig != "" {
	// 	env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	cmd := exec.Command(binary, args...)
	log.Info("Exec", "command", cmd.String())
	// cmd.Env = env
	out, err := cmd.CombinedOutput()
	log.Info(string(out))

	if err != nil {
		log.Error(err, "failed to execute InstallRelease() command", "command", cmd.String())
		err = fmt.Errorf("failed to install kubernetes resources of %s: %s", chartName, string(out))

		rollbackErr := DeleteReleaseIfExists(name, namespace)
		if rollbackErr != nil {
			log.Error(err, "failed to rollback installed helm release after InstallRelease() failure", "name", name, "namespace", namespace)
		}
		return err
	}

	return nil
}

// CheckRelease checks if the release with the given name and namespace exist.
func CheckRelease(name, namespace string) (exist bool, err error) {
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return exist, err
	}

	cmd := exec.Command(helmCmd[0], "status", name, "-n", namespace)
	// support multiple cluster management
	// if types.KubeConfig != "" {
	// 	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }

	resultBytes, err := cmd.CombinedOutput()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus := status.ExitStatus()
				log.V(1).Info("Exit", "Status", exitStatus)
				if exitStatus == 1 {
					err = nil
				}
			}
		} else {
			log.Error(err, "failed to execute CheckRelease() command", "command", cmd.String())
			return exist, err
		}
	} else {
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.ExitStatus() == 0 {
			// ####### helm status output ######
			// NAME: demo-dataset
			// LAST DEPLOYED: Thu Mar 16 17:08:06 2023
			// NAMESPACE: default
			// STATUS: deployed
			// REVISION: 1
			// TEST SUITE: None
			// ####### END #######
			resultLines := strings.Split(string(resultBytes), "\n")
			for _, line := range resultLines {
				if strings.HasPrefix(line, "STATUS: ") {
					if strings.Replace(line, "STATUS: ", "", 1) == "deployed" {
						exist = true
					} else {
						rollbackErr := DeleteRelease(name, namespace)
						if rollbackErr != nil {
							err = errors.Wrapf(rollbackErr, "failed to rollback failed release (namespace: %s, name: %s)", namespace, name)
						}
					}
				}
			}
		} else {
			if waitStatus.ExitStatus() != -1 {
				return exist, fmt.Errorf("unexpected return code %d when exec helm status %s -n %s",
					waitStatus.ExitStatus(),
					name,
					namespace)
			}
		}
	}

	return exist, err
}

// DeleteRelease deletes release with the name and namespace
func DeleteRelease(name, namespace string) error {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	args := []string{"uninstall", name, "-n", namespace}
	cmd := exec.Command(binary, args...)
	log.Info("Exec", "command", cmd.String())
	// env := os.Environ()
	// if types.KubeConfig != "" {
	// 	env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }
	// return syscall.Exec(cmd, args, env)
	out, err := cmd.Output()
	log.Info("delete release", "result", string(out))
	if err != nil {
		log.Error(err, "failed to execute DeleteRelease() command", "command", cmd.String())
		return fmt.Errorf("failed to delete engine-related kubernetes resources")
	}
	return nil
}

// ListReleases return an array with all releases' names in a given namespace
func ListReleases(namespace string) (releases []string, err error) {
	releases = []string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releases, err
	}

	cmd := exec.Command(helmCmd[0], "list", "-q", "-n", namespace)
	// support multiple cluster management
	// if types.KubeConfig != "" {
	// 	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }
	out, err := cmd.Output()
	if err != nil {
		return releases, err
	}
	return strings.Split(string(out), "\n"), nil
}

// ListReleaseMap returns a map with all releases' names and app versions in a given namespace.
func ListReleaseMap(namespace string) (releaseMap map[string]string, err error) {
	releaseMap = map[string]string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releaseMap, err
	}

	cmd := exec.Command(helmCmd[0], "list", "-n", namespace)
	// // support multiple cluster management
	// if types.KubeConfig != "" {
	// 	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }
	out, err := cmd.Output()
	if err != nil {
		return releaseMap, err
	}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.Trim(line, " ")
		if !strings.Contains(line, "NAME") {
			cols := strings.Fields(line)
			// log.Debugf("%d cols: %v", len(cols), cols)
			if len(cols) > 1 {
				// log.Debugf("releaseMap: %s=%s\n", cols[0], cols[len(cols)-1])
				releaseMap[cols[0]] = cols[len(cols)-1]
			}
		}
	}

	return releaseMap, nil
}

// ListAllReleasesWithDetail returns a map with all releases' names and other info in a given namespace
func ListAllReleasesWithDetail(namespace string) (releaseMap map[string][]string, err error) {
	releaseMap = map[string][]string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releaseMap, err
	}

	cmd := exec.Command(helmCmd[0], "list", "--all", "-n", namespace)
	// support multiple cluster management
	// if types.KubeConfig != "" {
	// 	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }
	out, err := cmd.Output()
	if err != nil {
		return releaseMap, err
	}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.Trim(line, " ")
		if !strings.Contains(line, "NAME") {
			cols := strings.Fields(line)
			// log.Debugf("%d cols: %v", len(cols), cols)
			if len(cols) > 3 {
				// log.Debugf("releaseMap: %s=%s\n", cols[0], cols)
				releaseMap[cols[0]] = cols
			}
		}
	}

	return releaseMap, nil
}

// DeleteReleaseIfExists deletes a release with given name and namespace if it exists.
// A wrapper of CheckRelease() and DeleteRelease()
func DeleteReleaseIfExists(name, namespace string) error {
	existed, err := CheckRelease(name, namespace)
	if err != nil {
		return err
	} else if existed {
		return DeleteRelease(name, namespace)
	}
	// release not found
	return nil
}
