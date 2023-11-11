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
package utils

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestGenUFSRootPathForAlluxio(t *testing.T) {
	testCases := map[string]struct {
		mounts       []datav1alpha1.Mount
		wantMount    *datav1alpha1.Mount
		wantRootPath string
	}{
		"test multi mount item case 1": {
			mounts:       mockMountMultiItems(mockMountSingleItem("spark", "local://mnt/local/path", "")),
			wantRootPath: "/underFSStorage",
			wantMount:    nil,
		},
		"test single mount item with fluid native scheme case 1": {
			mounts:       mockMountSingleItem("spark", "local://mnt/local/path", ""),
			wantRootPath: "/underFSStorage",
			wantMount:    nil,
		},
		"test single mount item with fluid native scheme case 2": {
			mounts:       mockMountSingleItem("spark", "pvc://mnt/local/path", ""),
			wantRootPath: "/underFSStorage",
			wantMount:    nil,
		},
		"test single mount item with mount path case 1": {
			mounts:       mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/mnt"),
			wantRootPath: "/underFSStorage",
			wantMount:    nil,
		},
		"test single mount item with mount path case 2": {
			mounts:       mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/"),
			wantRootPath: "http://fluid.io/spark/spark-3.0.2",
			wantMount:    &mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/")[0],
		},
		"test single mount item with mount path case 3": {
			mounts:       mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", ""),
			wantRootPath: "/underFSStorage",
			wantMount:    nil,
		},
	}

	for k, item := range testCases {
		gotRootPath, m := UFSPathBuilder{}.GenAlluxioUFSRootPath(item.mounts)
		if gotRootPath != item.wantRootPath {
			t.Errorf("%s check failure, want:%s, got:%s", k, item.wantRootPath, gotRootPath)
		}
		if !reflect.DeepEqual(m, item.wantMount) {
			t.Errorf("%s check mount failure, want:%v, got:%v", k, item.wantMount, m)
		}
	}
}

func TestGetAlluxioMountPath(t *testing.T) {
	testCases := map[string]struct {
		curMount datav1alpha1.Mount
		wantPath string
		mounts   []datav1alpha1.Mount
	}{
		"test only one mount item case 1": {
			curMount: mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "")[0],
			mounts:   mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", ""),
			wantPath: "/spark",
		}, "test only one mount item case 2": {
			curMount: mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/")[0],
			mounts:   mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/"),
			wantPath: "/",
		},
		"test only one mount item with define mount path case 1": {
			curMount: mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/mnt/user/define/path")[0],
			mounts:   mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/mnt/user/define/path"),
			wantPath: "/mnt/user/define/path",
		},
		"test multi mount items case 1": {
			curMount: mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "")[0],
			mounts:   mockMountMultiItems(mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "")),
			wantPath: "/spark",
		},
		"test multi mount items with define mount path case 1": {
			curMount: mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/mnt/user/define/path")[0],
			mounts:   mockMountMultiItems(mockMountSingleItem("spark", "http://fluid.io/spark/spark-3.0.2", "/mnt/user/define/path")),
			wantPath: "/mnt/user/define/path",
		},
	}

	for k, item := range testCases {
		gotPath := UFSPathBuilder{}.GenAlluxioMountPath(item.curMount)
		if gotPath != item.wantPath {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantPath, gotPath)
		}
	}
}

func mockMountSingleItem(name, mPoint, path string) []datav1alpha1.Mount {
	return []datav1alpha1.Mount{
		{
			Name:       name,
			MountPoint: mPoint,
			Path:       path,
		},
	}
}

func mockMountMultiItems(m []datav1alpha1.Mount) []datav1alpha1.Mount {
	return append(m,
		datav1alpha1.Mount{
			Name:       "spark-append-1",
			MountPoint: "https://fluid.io/apache/spark/spark-3.0.2/",
		},
		datav1alpha1.Mount{
			Name:       "flink-append-2",
			MountPoint: "https://fluid.io/apache/flink/flink-1.13.0/",
		},
	)
}
