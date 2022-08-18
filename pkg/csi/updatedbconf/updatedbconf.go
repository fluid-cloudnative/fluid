package updatedbconf

import (
	"fmt"
	"strings"
)

const (
	updatedbConfPath    = "/host-etc/updatedb.conf"
	configKeyPruneFs    = "PRUNEFS"
	configKeyPrunePaths = "PRUNEPATHS"
)

// updateLine add new config items to a line
func updateLine(line string, key string, values []string) string {
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
		return oldLine
	}
	current = append(current, newValues...)
	return fmt.Sprintf(`%s="%s"`, key, strings.Join(current, " "))
}

// updateConfig parse the updatedb.conf by line and add the `fs` `path` items
func updateConfig(content string, newFs []string, newPaths []string) (string, error) {
	lines := strings.Split(content, "\n")
	var hasPruneFsConfig = false
	var hasPrunepPathConfig = false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		// update PRUNEFS
		if strings.HasPrefix(line, configKeyPruneFs) {
			hasPruneFsConfig = true
			lines[i] = updateLine(line, configKeyPruneFs, newFs)
		}
		// update PRUNEPATHS
		if strings.HasPrefix(line, configKeyPrunePaths) {
			hasPrunepPathConfig = true
			lines[i] = updateLine(line, configKeyPrunePaths, newPaths)
		}
	}
	// no PRUNEFS or PRUNEPATHS in config file, append new config line
	if !hasPruneFsConfig && len(newFs) > 0 {
		lines = append(lines, fmt.Sprintf(`%s="%s"`, configKeyPruneFs, strings.Join(newFs, " ")))
	}
	if !hasPrunepPathConfig && len(newPaths) > 0 {
		lines = append(lines, fmt.Sprintf(`%s="%s"`, configKeyPrunePaths, strings.Join(newPaths, " ")))
	}
	newContent := strings.Join(lines, "\n")
	return newContent, nil
}
