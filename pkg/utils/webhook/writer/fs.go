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
