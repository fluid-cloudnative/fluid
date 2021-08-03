package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
)

// ShouldCheckUFS checks if it requires checking UFS
func (e *JindoEngine) ShouldCheckUFS() (should bool, err error) {
	should = true
	return
}

// PrepareUFS do all the UFS preparations
func (e *JindoEngine) PrepareUFS() (err error) {
	// For Jindo Engine, not need to prepare UFS
	return
}

// UsedStorageBytes returns used storage size of Jindo in bytes
func (e *JindoEngine) UsedStorageBytes() (value int64, err error) {

	return
}

// FreeStorageBytes returns free storage size of Jindo in bytes
func (e *JindoEngine) FreeStorageBytes() (value int64, err error) {
	return
}

// return total storage size of Jindo in bytes
func (e *JindoEngine) TotalStorageBytes() (value int64, err error) {
	return
}

// return the total num of files in Jindo
func (e *JindoEngine) TotalFileNums() (value int64, err error) {
	return
}

// report jindo summary
func (e *JindoEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// JindoEngine hasn't support UpdateOnUFSChange
func (e *JindoEngine) UpdateOnUFSChange() (updateReady bool, err error) {
	return
}
