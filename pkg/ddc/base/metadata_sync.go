package base

import (
	"strconv"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
)

// MetadataSyncResult describes result for asynchronous metadata sync
type MetadataSyncResult struct {
	Done      bool
	StartTime time.Time
	UfsTotal  string
	FileNum   string
	Err       error
}

// SafeClose closes the metadataSyncResultChannel but ignores panic when the channel is already closed.
// Returns true if the channel is already closed.
func SafeClose(ch chan MetadataSyncResult) (closed bool) {
	if ch == nil {
		return
	}
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	close(ch)
	return false
}

// SafeSend sends result to the metadataSyncResultChannel but ignores panic when the channel is already closed
// Returns true if the channel is already closed.
func SafeSend(ch chan MetadataSyncResult, result MetadataSyncResult) (closed bool) {
	if ch == nil {
		return
	}
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	ch <- result
	return false
}

func RecordDatasetMetrics(result MetadataSyncResult, datasetNamespace, datasetName string, log logr.Logger) {
	if len(result.UfsTotal) != 0 {
		if ufsTotal, ignoredErr := utils.FromHumanSize(result.UfsTotal); ignoredErr == nil {
			metrics.GetDatasetMetrics(datasetNamespace, datasetName).SetUFSTotalSize(float64(ufsTotal))
		} else {
			log.Error(ignoredErr, "fail to parse result.UfsTotal", "result.UfsTotal", result.UfsTotal)
		}
	}

	if len(result.FileNum) != 0 {
		if fileNum, ignoredErr := strconv.Atoi(result.FileNum); ignoredErr == nil {
			metrics.GetDatasetMetrics(datasetNamespace, datasetName).SetUFSFileNum(float64(fileNum))
		} else {
			log.Error(ignoredErr, "fail to atoi result.FileNum", "result.FileNum", result.FileNum)
		}
	}
}
