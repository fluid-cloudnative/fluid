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

package utils

import "regexp"

// SplitSchemaAddr splits the address string into 2 parts: proto and addr.
// Proto is set to "tcp" in default.
// It returns proto and addr separately.
func SplitSchemaAddr(addr string) (string, string) {
	parts := reSchema.FindStringSubmatch(addr)
	proto, addr := parts[1], parts[2]
	if proto == "" {
		proto = "tcp"
	}
	return proto, addr
}

var reSchema = regexp.MustCompile("^(?:([a-z0-9]+)://)?(.*)$")
