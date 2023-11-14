/*
Copyright 2021 The Fluid Authors.

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

package writer

import (
	"fmt"
	"os"
	"path"

	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/generator"
)

func WriteCertsToDir(dir string, certs *generator.Artifacts) error {

	_, err := os.Stat(dir)
	switch {
	case os.IsNotExist(err):
		log.Info("cert directory doesn't exist, creating", "directory", dir)
		err = os.MkdirAll(dir, 0640)
		if err != nil {
			return fmt.Errorf("can't create dir: %v", dir)
		}
	case err != nil:
		return err
	}

	// write or overwrite the cert for ca
	filename := path.Join(dir, CAKeyName)
	if err := os.WriteFile(filename, certs.CAKey, 0600); err != nil {
		return err
	}
	filename = path.Join(dir, CACertName)
	if err := os.WriteFile(filename, certs.CACert, 0600); err != nil {
		return err
	}

	// write or overwrite the certs for serverName
	filename = path.Join(dir, ServerCertName)
	if err := os.WriteFile(filename, certs.Cert, 0600); err != nil {
		return err
	}
	filename = path.Join(dir, ServerKeyName)
	if err := os.WriteFile(filename, certs.Key, 0600); err != nil {
		return err
	}

	// write or overwrite the certs for serverName2
	filename = path.Join(dir, ServerCertName2)
	if err := os.WriteFile(filename, certs.Cert, 0600); err != nil {
		return err
	}
	filename = path.Join(dir, ServerKeyName2)
	if err := os.WriteFile(filename, certs.Key, 0600); err != nil {
		return err
	}
	return nil
}
