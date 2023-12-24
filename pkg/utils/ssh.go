/*
Copyright 2023 The Fluid Authors.

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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"golang.org/x/crypto/ssh"
)

func GenerateSSHConfig(releaseName string, parallelism int32) (*common.SSHConfig, error) {
	bitSize := 2048

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// id_rsa
	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(&privBlock)

	// id_rsa.pub
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	sshConfig := &common.SSHConfig{
		PrivateKey:     string(privateKeyBytes),
		PublicKey:      string(publicKeyBytes),
		AuthorizedKeys: string(publicKeyBytes),
		Config:         generateHostsConfig(releaseName, parallelism),
	}
	return sshConfig, nil
}

func generateHostsConfig(releaseName string, parallelism int32) string {
	config := "StrictHostKeyChecking no\nUserKnownHostsFile /dev/null\n"

	for i := 0; i < int(parallelism); i++ {
		// Host name format is ${releaseName}-workers-${completionIndex}
		hostName := fmt.Sprintf("%s-workers-%d", releaseName, i)
		// Service Name is ${jobName}-svc, keep consistent with helm yaml.
		subdomain := fmt.Sprintf("%s-svc", releaseName)

		// Note: this format will be parsed to get the $host.$domain for ping.
		// grep "Host " config | cut -d " " -f 2
		config += "Host " + hostName + "\n"
		config += "  HostName " + hostName + "." + subdomain + "\n"
	}

	return config
}
