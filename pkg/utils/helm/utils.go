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
	"strings"
	"syscall"
)

/**
* install the release with cmd: helm install -f values.yaml chart_name, support helm v3
 */
func InstallRelease(name string, namespace string, valueFile string, chartName string) error {
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
	log.V(1).Info("Exec", "args", args)

	// env := os.Environ()
	// if types.KubeConfig != "" {
	// 	env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	cmd := exec.Command(binary, args...)
	// cmd.Env = env
	out, err := cmd.CombinedOutput()
	log.Info(string(out))

	if err != nil {
		log.Error(err, "failed to execute", "args", strings.Join(args, " "))
	}

	return err
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

	if err := cmd.Start(); err != nil {
		// log.Fatalf("cmd.Start: %v", err)
		// log.Error(err)
		log.Error(err, "failed to execute")
		return exist, err
	}

	err = cmd.Wait()
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
			log.Error(err, "cmd.Wait")
			return exist, err
		}
	} else {
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.ExitStatus() == 0 {
			exist = true
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

/*
* Delete Release given name and namespace
 */
func DeleteRelease(name, namespace string) error {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	args := []string{"uninstall", name, "-n", namespace}
	cmd := exec.Command(binary, args...)

	// env := os.Environ()
	// if types.KubeConfig != "" {
	// 	env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }
	// return syscall.Exec(cmd, args, env)
	out, err := cmd.Output()
	log.V(1).Info("delete release", "result", string(out))
	return err
}

/*
* return an array with all releases' names in a given namespace
 */
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

// ListAllReleasesWithDetail return a map with all releases' names and other info in a given namespace
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

/*
* Delete a release with given name and namespace if it exists.
* A wrapper of CheckRelease() and DeleteRelease()
 */
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
