/*
Copyright 2024 The Fluid Authors.

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

package operations

import "strings"

// DecommissionWorkers signals the Alluxio master to decommission the given
// workers. Each address must be in "<host>:<rpcPort>" form.
// The call is idempotent: re-issuing it against an already-decommissioned
// worker is safe.
func (a AlluxioFileUtils) DecommissionWorkers(addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}
	command := []string{
		"alluxio", "fsadmin", "decommission",
		"--addresses", strings.Join(addresses, ","),
	}
	_, _, err := a.exec(command, false)
	if err != nil {
		a.log.Error(err, "AlluxioFileUtils.DecommissionWorkers() failed", "addresses", addresses)
	}
	return err
}

// CountActiveWorkers returns the number of workers currently tracked by the
// Alluxio master according to "alluxio fsadmin report capacity".
func (a AlluxioFileUtils) CountActiveWorkers() (int, error) {
	report, _, err := a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, false)
	if err != nil {
		a.log.Error(err, "AlluxioFileUtils.CountActiveWorkers() failed")
		return 0, err
	}
	return parseActiveWorkerCount(report), nil
}

// parseActiveWorkerCount counts workers in the capacity report produced by
// "alluxio fsadmin report capacity". Worker entries begin at the non-indented
// line after the "Worker Name" header; the indented line that follows each
// entry contains the used-capacity detail.
//
//	Worker Name      Last Heartbeat   Storage       MEM
//	192.168.1.147    0                capacity      2048.00MB    <- worker entry
//	                                 used          443.89MB (21%) <- detail, indented
//	192.168.1.146    0                capacity      2048.00MB    <- worker entry
//	                                 used          0B (0%)
func parseActiveWorkerCount(report string) int {
	inWorkerSection := false
	count := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "Worker Name") {
			inWorkerSection = true
			continue
		}
		if !inWorkerSection || strings.TrimSpace(line) == "" {
			continue
		}
		// Non-indented lines are new worker entries; indented lines are
		// the used-capacity continuation for the previous entry.
		if line[0] != ' ' && line[0] != '\t' {
			count++
		}
	}
	return count
}
