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

package base

import (
	"errors"
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

// RecordDatasetMetrics records dataset-related metrics from the given MetadataSyncResult
func RecordDatasetMetrics(result MetadataSyncResult, datasetNamespace, datasetName string, log logr.Logger) {
	if len(datasetNamespace) == 0 {
		argErr := errors.New("invalid argument: datasetNamespace should not be empty")
		log.Error(argErr, "fail to validate RecordDatasetMetrics arguments")
		return
	}

	if len(datasetName) == 0 {
		argErr := errors.New("invalid argument: datasetName should not be empty")
		log.Error(argErr, "fail to validate RecordDatasetMetrics arguments")
		return
	}

	if len(result.UfsTotal) != 0 {
		if ufsTotal, parseErr := utils.FromHumanSize(result.UfsTotal); parseErr == nil {
			metrics.GetDatasetMetrics(datasetNamespace, datasetName).SetUFSTotalSize(float64(ufsTotal))
		} else {
			log.Error(parseErr, "fail to parse result.UfsTotal", "result.UfsTotal", result.UfsTotal)
		}
	}

	if len(result.FileNum) != 0 {
		if fileNum, parseErr := strconv.Atoi(result.FileNum); parseErr == nil {
			metrics.GetDatasetMetrics(datasetNamespace, datasetName).SetUFSFileNum(float64(fileNum))
		} else {
			log.Error(parseErr, "fail to atoi result.FileNum", "result.FileNum", result.FileNum)
		}
	}
}
