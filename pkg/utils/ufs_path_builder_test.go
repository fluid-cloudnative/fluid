/*
Copyright 2023 The Fluid Author.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
