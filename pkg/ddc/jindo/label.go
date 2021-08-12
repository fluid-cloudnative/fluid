package jindo

import "github.com/fluid-cloudnative/fluid/pkg/common"

func (e *JindoEngine) getCommonLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name
}

func (e *JindoEngine) getFuseLabelname() string {
	return common.LabelAnnotationFusePrefix + e.namespace + "-" + e.name
}
