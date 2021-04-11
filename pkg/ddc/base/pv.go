package base

import "fmt"

func (info *RuntimeInfo) GetPersistentVolumeName() string {
	pvName := fmt.Sprintf("%s-%s", info.GetNamespace(), info.GetName())
	if info.IsDeprecatedPVName() {
		pvName = info.GetName()
	}

	return pvName
}
