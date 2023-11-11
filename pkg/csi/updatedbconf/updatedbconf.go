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

package updatedbconf

import (
	"fmt"
	"strings"
)

const (
	updatedbConfPath       = "/host-etc/updatedb.conf"
	updatedbConfBackupPath = "/host-etc/updatedb.conf.backup"
	configKeyPruneFs       = "PRUNEFS"
	configKeyPrunePaths    = "PRUNEPATHS"
	modifiedByFluidComment = "# Modified by Fluid"
)

// updateLine add new config items to a line
// return false if the config line has no changes
func updateLine(line string, key string, values []string) (string, bool) {
	oldLine := line
	line = strings.TrimPrefix(line, key)
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "=")
	line = strings.TrimSpace(line)
	line = strings.Trim(line, `"`)
	line = strings.TrimSpace(line)
	current := strings.Split(line, " ")

	newValues := []string{}
	for _, v := range values {
		exist := false
		for _, s := range current {
			if v == s {
				exist = true
				break
			}
		}
		if !exist && v != "" {
			newValues = append(newValues, v)
		}
	}
	// no new items, skip update
	if len(newValues) == 0 {
		return oldLine, false
	}
	current = append(current, newValues...)
	return fmt.Sprintf(`%s="%s"`, key, strings.Join(current, " ")), true
}

// updateConfig parse the updatedb.conf by line and add the `fs` `path` items
func updateConfig(content string, newFs []string, newPaths []string) (string, error) {
	lines := strings.Split(content, "\n")
	var hasPruneFsConfig = false
	var hasPrunepPathConfig = false
	var configChange = false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		// update PRUNEFS
		if strings.HasPrefix(line, configKeyPruneFs) {
			hasPruneFsConfig = true
			if newline, shouldUpdate := updateLine(line, configKeyPruneFs, newFs); shouldUpdate {
				configChange = true
				lines[i] = newline
			}
		}
		// update PRUNEPATHS
		if strings.HasPrefix(line, configKeyPrunePaths) {
			hasPrunepPathConfig = true
			if newline, shouldUpdate := updateLine(line, configKeyPrunePaths, newPaths); shouldUpdate {
				configChange = true
				lines[i] = newline
			}
		}
	}
	// no PRUNEFS or PRUNEPATHS in config file, append new config line
	if !hasPruneFsConfig && len(newFs) > 0 {
		configChange = true
		lines = append(lines, fmt.Sprintf(`%s="%s"`, configKeyPruneFs, strings.Join(newFs, " ")))
	}
	if !hasPrunepPathConfig && len(newPaths) > 0 {
		configChange = true
		lines = append(lines, fmt.Sprintf(`%s="%s"`, configKeyPrunePaths, strings.Join(newPaths, " ")))
	}
	// if the config file already has the expected items, return the original content directly
	if !configChange {
		return content, nil
	}
	newContent := strings.Join(lines, "\n")
	return newContent, nil
}
