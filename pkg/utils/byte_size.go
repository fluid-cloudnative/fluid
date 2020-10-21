package utils

import (
	"fmt"
	"github.com/docker/go-units"
	"regexp"
	"strconv"
	"strings"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
)

type unitMap map[string]int64

var (
	binaryMap = unitMap{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
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
// size with 1024 as multiplier (eg. "44kB" = "(44 * 1024)B").
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
