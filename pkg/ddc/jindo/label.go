package jindo

import "github.com/fluid-cloudnative/fluid/pkg/common"

func (e *JindoEngine) getStorageLabelname(read common.ReadType, storage common.StorageType) string {
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
