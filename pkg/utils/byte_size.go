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

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/go-units"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
	EiB = 1024 * PiB
)

type unitMap map[string]int64

var (
	binaryMap = unitMap{
		"k": KiB,
		"m": MiB,
		"g": GiB,
		"t": TiB,
		"p": PiB,
		"e": EiB,
	}
	sizeRegex = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[iI]?[bB]?$`)
)

var binaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

// BytesSize returns a human-readable size in bytes, kibibytes,
// mebibytes, gibibytes, or tebibytes, but with a B, kB, MB unit style.
// This is to make byte units be in consistent with Alluxio
// See https://github.com/Alluxio/alluxio/blob/master/core/common/src/main/java/alluxio/util/FormatUtils.java#L135
func BytesSize(size float64) string {
	return units.CustomSize("%.2f%s", size, 1024.0, binaryAbbrs)
}

// FromHumanSize returns an integer from a human-readable specification of a
// size with 1024 as multiplier
// e.g.:
//  1. 1 KiB = 1024 byte
func FromHumanSize(size string) (int64, error) {
	return parseSize(size, binaryMap)
}

func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[3])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= float64(mul)
	}

	return int64(size), nil
}
