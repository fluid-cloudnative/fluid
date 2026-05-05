/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutils "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// CacheFileUtil defines the interface for cache system operations
type CacheFileUtil interface {
	Mount(command []string, timeout time.Duration) (stdout string, err error)
}

type CacheFileUtilImpl struct {
	podName   string
	namespace string
	container string
	log       logr.Logger
}

func NewCacheFileUtil(podName string, containerName string, namespace string, log logr.Logger) CacheFileUtil {

	return &CacheFileUtilImpl{
		podName:   podName,
		namespace: namespace,
		container: containerName,
		log:       log,
	}
}

// exec with timeout
func (c *CacheFileUtilImpl) exec(command []string, timeout time.Duration) (stdout string, stderr string, err error) {
	// redact sensitive info in command for printing
	redactedCommand := securityutils.FilterCommand(command)

	c.log.V(1).Info("Exec command start", "command", redactedCommand)
	stdout, stderr, err = kubeclient.ExecCommandInContainerWithTimeout(c.podName, c.container, c.namespace, command, timeout)
	if err != nil {
		err = errors.Wrapf(err, "error when executing command %v", redactedCommand)
		return
	}
	c.log.V(1).Info("Exec command finished", "command", redactedCommand)

	return
}

func (c *CacheFileUtilImpl) Mount(command []string, timeout time.Duration) (stdout string, err error) {
	stdout, stderr, err := c.exec(command, timeout)

	if err != nil {
		c.log.Error(err, "CacheFileUtils.Mount() failed", "stdout", stdout, "stderr", stderr)
		return
	}

	return stdout, nil
}
