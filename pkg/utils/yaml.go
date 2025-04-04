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

	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write(data)
	if err != nil {
		log.Error(err, "failed to write file", "data", data, "fileName", file.Name())
	}
	return err
}
