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

package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils/cmdguard"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("kubectl")
}

// CreateConfigMapFromFile creates configMap from file.
func CreateConfigMapFromFile(name string, key, fileName string, namespace string) (err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"create", "configmap", name,
		"--namespace", namespace,
		fmt.Sprintf("--from-file=%s=%s", key, fileName)}

	out, err := kubectl(args)
	log.V(1).Info("exec: ", "cmd", args)
	log.V(1).Info(fmt.Sprintf("result: %s", string(out)))
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to execute %v", args))
	}

	return
}

// kubectl executes command with arguments (string array)
func kubectl(args []string) ([]byte, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return nil, err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.V(1).Info("exec", "binary", binary, "cmd", strings.Join(args, " "))
	// env := os.Environ()
	// if types.KubeConfig != "" {
	// 	env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	// }

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd, err := cmdguard.Command(binary, args...)
	if err != nil {
		return nil, err
	}
	// cmd.Env = env
	return cmd.CombinedOutput()
}
