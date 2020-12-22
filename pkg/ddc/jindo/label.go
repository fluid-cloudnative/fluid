package jindo
import "github.com/fluid-cloudnative/fluid/pkg/common"

type readType string

const (
	humanReadType readType = "human-"

	// rawReadType readType = "raw-"
)

type storageType string

const (
	memoryStorageType storageType = "mem-"

	diskStorageType storageType = "disk-"

	totalStorageType storageType = "total-"
)

func (e *JindoEngine) getStoragetLabelname(read readType, storage storageType) string {
	return common.LabelAnnotationStorageCapacityPrefix +
		string(read) +
		e.runtimeType +
		"-" +
		string(storage) +
		e.namespace +
		"-" +
		e.name
}

func (e *JindoEngine) getCommonLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name
}

func (e *JindoEngine) getRuntimeLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.namespace + "-" + e.name
}
