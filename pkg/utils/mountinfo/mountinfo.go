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

package mountinfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Mount struct {
	Subtree        string
	MountPath      string
	FilesystemType string
	PeerGroups     map[int]bool
	ReadOnly       bool
	Count          int
}

// Parse one line of /proc/self/mountinfo.
//
// The line contains the following space-separated fields:
//
//	[0] mount ID
//	[1] parent ID
//	[2] major:minor
//	[3] root
//	[4] mount point
//	[5] mount options
//	[6...n-1] optional field(s)
//	[n] separator
//	[n+1] filesystem type
//	[n+2] mount source
//	[n+3] super options
//
// For more details, see https://www.kernel.org/doc/Documentation/filesystems/proc.txt
func parseMountInfoLine(line string) *Mount {
	fields := strings.Split(line, " ")
	if len(fields) < 10 {
		return nil
	}

	var mnt = &Mount{}
	mnt.Subtree = unescapeString(fields[3])
	mnt.MountPath = unescapeString(fields[4])
	for _, opt := range strings.Split(fields[5], ",") {
		if opt == "ro" {
			mnt.ReadOnly = true
		}
	}
	// Count the optional fields.  In case new fields are appended later,
	// don't simply assume that n == len(fields) - 4.
	n := 6
	mnt.PeerGroups = map[int]bool{}
	for ; n < len(fields) && fields[n] != "-"; n++ {
		if peerGroupTag, peerGroup, err := peerGroupFromString(fields[n]); err != nil {
			glog.V(0).Infof("WARNING: fail to parse peer group info from mount point %s's option %s: %v", mnt.MountPath, fields[n], err)
			return nil
		} else if peerGroupTag == "shared" || peerGroupTag == "master" {
			mnt.PeerGroups[peerGroup] = true
		}
	}
	if n >= len(fields) {
		glog.V(0).Infof("WARNING: invalid mount info line with no separator: %s", line)
		return nil
	}

	// n now equals to the index of the separator("-") in the mount info line
	if n+3 >= len(fields) {
		glog.V(0).Infof("WARNING: invalid mount info line without enough fields: %s", line)
		return nil
	}
	// if n > 6 {
	// 	if shared, peerGroup, err := peerGroupFromString(fields[6]); err != nil {
	// 		return nil
	// 	} else if shared {
	// 		mnt.PeerGroup = &peerGroup
	// 	}
	// }
	mnt.FilesystemType = unescapeString(fields[n+1])
	mnt.Count = 1
	return mnt
}

func readMountInfo(r io.Reader) (map[string]*Mount, error) {
	mountsByPath := make(map[string]*Mount)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		mnt := parseMountInfoLine(line)
		if mnt == nil {
			glog.V(5).Infof("ignoring invalid mountinfo line %q", line)
			continue
		}

		// We can only use mountpoints that are directories for fluid.
		if len(mnt.PeerGroups) == 0 {
			glog.V(6).Infof("ignoring mountpoint %q because it is not rshared or rslave", mnt.MountPath)
			continue
		}

		if oldMnt, ok := mountsByPath[mnt.MountPath]; ok {
			// record mountpoint count in mountinfo
			mnt.Count = oldMnt.Count + 1
		}
		// Note this overrides the info if we have seen the mountpoint
		// earlier in the file. This is correct behavior because the
		// mountpoints are listed in mount order.
		mountsByPath[mnt.MountPath] = mnt
	}
	return mountsByPath, nil
}

// loadMountInfo populates the Mount mappings by parsing /proc/self/mountinfo.
// It returns an error if the Mount mappings cannot be populated.
func loadMountInfo() (map[string]*Mount, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	mountsByPath, err := readMountInfo(file)
	if err != nil {
		return nil, err
	}
	return mountsByPath, nil
}

// Unescape octal-encoded escape sequences in a string from the mountinfo file.
// The kernel encodes the ' ', '\t', '\n', and '\\' bytes this way.  This
// function exactly inverts what the kernel does, including by preserving
// invalid UTF-8.
func unescapeString(str string) string {
	var sb strings.Builder
	for i := 0; i < len(str); i++ {
		b := str[i]
		if b == '\\' && i+3 < len(str) {
			if parsed, err := strconv.ParseInt(str[i+1:i+4], 8, 8); err == nil {
				b = uint8(parsed)
				i += 3
			}
		}
		sb.WriteByte(b)
	}
	return sb.String()
}

func peerGroupFromString(str string) (peerGroupTag string, peerGroup int, err error) {
	peerGroupTag, peerGroup = "", -1

	fields := strings.Split(str, ":")
	if len(fields) != 2 {
		err = fmt.Errorf("%s is not a peer group tag", str)
		return
	}
	if fields[0] != "shared" && fields[0] != "master" {
		err = fmt.Errorf("%s is not a legal peer group tag", fields[0])
		return
	}
	peerGroupTag = fields[0]

	peerGroup, err = strconv.Atoi(fields[1])
	if err != nil {
		err = errors.Wrapf(err, "failed to atoi %s", fields[1])
		return
	}

	return
}
