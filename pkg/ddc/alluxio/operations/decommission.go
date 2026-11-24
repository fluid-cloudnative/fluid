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

package operations

import (
	"fmt"
	"strings"
)

// DecommissionWorkers signals the Alluxio master to decommission the given
// workers. Each address must be in "<host>:<webPort>" form.
// The call is idempotent: re-issuing it against an already-decommissioned
// worker is safe.
//
// Requires Alluxio >= 2.9, where "fsadmin decommissionWorker" was introduced;
// against older masters this subcommand does not exist and the command exits
// non-zero. The caller (drainScalingDownWorkers/SyncReplicas) does not treat
// that as fatal: it retries on every reconcile, bounded by
// defaultWorkerDecommissionDeadline, after which it degrades to an
// ungraceful scale-down rather than stalling forever - so an old master
// still converges, just without the graceful drain.
func (a AlluxioFileUtils) DecommissionWorkers(addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}
	command := []string{
		"alluxio", "fsadmin", "decommissionWorker",
		"--addresses", strings.Join(addresses, ","),
		// --wait defaults to 5m, which would block this exec call (and the
		// engine's reconcile loop) waiting for the worker to idle. The
		// engine already polls CountActiveWorkers across reconciles, so
		// just initiate the decommission and return immediately.
		"--wait", "0s",
	}
	stdout, stderr, err := a.exec(command, false)
	if err != nil {
		a.log.Error(err, "AlluxioFileUtils.DecommissionWorkers() failed", "addresses", addresses, "stdout", stdout, "stderr", stderr)
		return err
	}
	// decommissionWorker can report a per-worker problem (an address it
	// couldn't resolve, a worker it couldn't reach) in its stdout while still
	// exiting 0 for the batch as a whole. stderr is deliberately not scanned
	// here: Alluxio's CLI is a JVM process and can print benign warnings to
	// stderr on a successful run (as every other command wrapper in this
	// package already assumes by only inspecting stderr once err != nil), so
	// treating any stderr output as failure would false-positive constantly.
	// The actual "did it work" signal is re-verified independently on every
	// subsequent reconcile via CountActiveWorkers, and a false positive here
	// only costs one extra bounded retry - so this stdout scan is a
	// best-effort signal, not the authoritative check.
	if reason, failed := decommissionOutputIndicatesFailure(stdout); failed {
		err = fmt.Errorf("decommissionWorker for addresses %v reported a possible failure (%s): stdout=%q stderr=%q",
			addresses, reason, stdout, stderr)
		a.log.Error(err, "AlluxioFileUtils.DecommissionWorkers() output looked unsuccessful", "addresses", addresses)
		return err
	}
	return nil
}

// decommissionOutputIndicatesFailure does a best-effort scan of
// decommissionWorker's stdout for signs that it didn't fully succeed despite
// exiting 0.
func decommissionOutputIndicatesFailure(stdout string) (reason string, failed bool) {
	lower := strings.ToLower(stdout)
	for _, indicator := range []string{"failed to", "unrecognized"} {
		if strings.Contains(lower, indicator) {
			return fmt.Sprintf("stdout contains %q", indicator), true
		}
	}
	return "", false
}

// CountActiveWorkers returns the number of live workers according to
// "alluxio fsadmin report capacity -live". The "-live" flag is what makes
// this safe to compare against immediately after a decommission: it asks the
// master for currently live workers rather than every worker it still has a
// record of, so a worker that was just decommissioned doesn't linger in the
// count until its heartbeat times out.
func (a AlluxioFileUtils) CountActiveWorkers() (int, error) {
	report, _, err := a.exec([]string{"alluxio", "fsadmin", "report", "capacity", "-live"}, false)
	if err != nil {
		a.log.Error(err, "AlluxioFileUtils.CountActiveWorkers() failed")
		return 0, err
	}
	count, err := parseActiveWorkerCount(report)
	if err != nil {
		a.log.Error(err, "AlluxioFileUtils.CountActiveWorkers() failed to parse capacity report", "report", report)
		return 0, err
	}
	return count, nil
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
//
// It returns an error rather than silently reporting 0 workers when the
// "Worker Name" header is never found: a report this caller can't recognize
// (a format change, an unexpected error message in place of the report, ...)
// must not be misread as "every worker has drained", since that's exactly
// the signal the engine acts on to let a scale-down proceed.
func parseActiveWorkerCount(report string) (int, error) {
	inWorkerSection := false
	count := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Worker Name") {
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
	if !inWorkerSection {
		return 0, fmt.Errorf("unrecognized capacity report format: missing %q header", "Worker Name")
	}
	return count, nil
}
