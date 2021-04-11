package jindo

import "github.com/fluid-cloudnative/fluid/pkg/common"

func (e *JindoEngine) getCommonLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name
}
