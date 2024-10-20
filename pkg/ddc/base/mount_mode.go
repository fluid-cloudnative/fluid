package base

import (
	"fmt"
	"strings"
)

type MountMode string

// Supported mount modes
const (
	MountPodMountMode MountMode = "MountPod"
	SidecarMountMode  MountMode = "Sidecar"
)

var SupportedMountModes = []MountMode{MountPodMountMode, SidecarMountMode}

type mountModeSelector map[MountMode]bool

func (mms mountModeSelector) Selected(mode MountMode) bool {
	_, exists := mms[mode]
	return exists
}

const (
	MountModeSelectAll  = "All"
	MountModeSelectNone = "None"
)

func ParseMountModeSelectorFromStr(mountModeStr string) (mountModeSelector, error) {
	ret := mountModeSelector{}
	if len(mountModeStr) == 0 {
		return ret, nil
	}

	mountModes := strings.Split(mountModeStr, ",")
	for _, mountMode := range mountModes {
		switch mountMode {
		case MountModeSelectAll:
			for _, mountMode := range SupportedMountModes {
				ret[mountMode] = true
			}
			return ret, nil
		case MountModeSelectNone:
			return ret, nil
		case string(MountPodMountMode):
			ret[MountPodMountMode] = true
		case string(SidecarMountMode):
			ret[SidecarMountMode] = true
		default:
			return nil, fmt.Errorf("unsupported mount mode %s, supported mount modes are %v", mountMode, SupportedMountModes)
		}
	}

	return ret, nil
}
