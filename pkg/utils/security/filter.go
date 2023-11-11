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

package security

import "strings"

var sensitiveKeys map[string]bool = map[string]bool{
	"aws.secretKey":          true,
	"aws.accessKeyId":        true,
	"fs.oss.accessKeyId":     true,
	"fs.oss.accessKeySecret": true,
}

func FilterCommand(command []string) (filteredCommand []string) {
	for _, str := range command {
		filteredCommand = append(filteredCommand, FilterString(str))
	}

	return
}

func FilterString(line string) string {
	for s := range sensitiveKeys {
		// if the log line contains a secret value redact it
		if strings.Contains(line, s) {
			line = s + "=[ redacted ]"
		}
	}

	return line
}

func UpdateSensitiveKey(key string) {
	if _, found := sensitiveKeys[key]; !found {
		sensitiveKeys[key] = true
	}
}
