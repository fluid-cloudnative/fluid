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

package disk

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type filesystemStat struct {
	device     string
	mountPoint string
	fsType     string
	size       float64
	avail      float64
}

func getFilesystemStats() ([]filesystemStat, error) {
	cmd := exec.Command("nsenter", "-m/proc/1/ns/mnt", "--", "df", "-B", "512", "-T")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	fsStats, err := parseFilesystemInfo(string(out))
	if err != nil {
		return nil, err
	}

	return fsStats, nil
}

func parseFilesystemInfo(mountInfos string) ([]filesystemStat, error) {
	var fsStats []filesystemStat
	mountInfoItems := strings.Split(mountInfos, "\n")
	for _, item := range mountInfoItems {
		// Ignore header
		if strings.Contains(item, "Use%") || item == "" {
			continue
		}

		parts := strings.Fields(item)

		if len(parts) < 6 {
			return nil, fmt.Errorf("malformed mount point information: %q", item)
		}

		toAppend := filesystemStat{
			device:     parts[0],
			fsType:     parts[1],
			mountPoint: parts[6],
		}

		totalBlocks, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, fmt.Errorf("error: %v when parsing mount point information: %q", err, item)
		}
		availBlocks, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			return nil, fmt.Errorf("error: %v when parsing mount point information: %q", err, item)
		}

		toAppend.size = totalBlocks * float64(512)
		toAppend.avail = availBlocks * float64(512)

		fsStats = append(fsStats, toAppend)
	}

	return fsStats, nil
}
