/*
Copyright 2020 The Fluid Authors.

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

package operations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	NOT_EXIST      = "not-exist"
	OTHER_ERR      = "other-err"
	FINE           = "fine"
	EXEC_ERR       = "exec-err"
	NEGATIVE_RES   = "negative-res"
	TOO_MANY_LINES = "too many lines"
	DATA_NUM       = "data nums not match"
	PARSE_ERR      = "parse err"
)

var _ = Describe("AlluxioFileUtils.NewAlluxioFileUtils", func() {
	It("should create AlluxioFileUtils correctly", func() {
		var expectedResult = AlluxioFileUtils{
			podName:   "hbase",
			namespace: "default",
			container: "hbase-container",
			log:       fake.NullLogger(),
		}
		result := NewAlluxioFileUtils("hbase", "hbase-container", "default", fake.NullLogger())
		Expect(result).To(Equal(expectedResult))
	})
})

var _ = Describe("AlluxioFileUtils.LoadMetaData", func() {
	var (
		patch *gomonkey.Patches
	)

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	It("should not return error", func() {
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))

		mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
			return "", "", nil
		}

		patch = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)

		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.LoadMetaData("/", true)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("AlluxioFileUtils.Du", func() {
	var (
		patch *gomonkey.Patches
	)

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when kubeclient exec returns error", func() {
		BeforeEach(func() {
			out1, out2, out3 := 111, 222, "%233"
			mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
				if strings.Contains(p4[4], EXEC_ERR) {
					return "does not exist", "", errors.New("exec-error")
				} else if strings.Contains(p4[4], TOO_MANY_LINES) {
					return "1\n2\n3\n4\n", "1\n2\n3\n4\n", nil
				} else if strings.Contains(p4[4], DATA_NUM) {
					return "1\n2\t3", "1\n2\t3", nil
				} else if strings.Contains(p4[4], PARSE_ERR) {
					return "1\n1\tdududu\tbbb\t", "1\n1\t2\tbbb\t", nil
				} else {
					return fmt.Sprintf("first line!\n%d\t%d\t(%s)\t2333", out1, out2, out3), "", nil
				}
			}

			patch = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)
		})

		It("should return error", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(EXEC_ERR)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when too many lines", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(TOO_MANY_LINES)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when data num not match", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(DATA_NUM)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when parse failed", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(PARSE_ERR)
			Expect(err).To(HaveOccurred())
		})

		It("should not return error", func() {
			o1, o2, o3, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(FINE)
			Expect(err).NotTo(HaveOccurred())
			Expect(o1).To(Equal(int64(111)))
			Expect(o2).To(Equal(int64(222)))
			Expect(o3).To(Equal("%233"))
		})
	})
})

var _ = Describe("AlluxioFileUtils.ReportSummary", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, err := a.ReportSummary()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			_, err := a.ReportSummary()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.LoadMetadataWithoutTimeout", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecWithoutTimeoutErr)
		})

		It("should return an error", func() {
			err := a.LoadMetadataWithoutTimeout("/")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecWithoutTimeoutCommon)
		})

		It("should not return an error", func() {
			err := a.LoadMetadataWithoutTimeout("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.LoadMetaData", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)
		})

		It("should return an error", func() {
			err := a.LoadMetaData("/", true)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			err := a.LoadMetaData("/", false)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.QueryMetaDataInfoIntoFile", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)
		})

		It("should return an error", func() {
			keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
			for _, keySet := range keySets {
				_, err := a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
			for _, keySet := range keySets {
				_, err := a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})

var _ = Describe("AlluxioFileUtils.Mkdir", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)
		})

		It("should return an error", func() {
			err := a.Mkdir("/")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "alluxio mkdir success", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			err := a.Mkdir("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.Mount", func() {
	var (
		a     AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = AlluxioFileUtils{}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)
		})

		It("should return an error", func() {
			testCases := []struct {
				readOnly bool
				shared   bool
				options  map[string]string
			}{
				{
					readOnly: true,
					shared:   true,
					options: map[string]string{
						"testKey": "testValue",
					},
				},
				{
					readOnly: true,
					shared:   false,
				},
				{
					readOnly: false,
					shared:   true,
				},
				{
					readOnly: false,
					shared:   false,
				},
			}

			for _, test := range testCases {
				err := a.Mount("/", "/", nil, test.readOnly, test.shared)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "alluxio mkdir success", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			testCases := []struct {
				readOnly bool
				shared   bool
				options  map[string]string
			}{
				{
					readOnly: true,
					shared:   true,
					options: map[string]string{
						"testKey": "testValue",
					},
				},
				{
					readOnly: true,
					shared:   false,
				},
				{
					readOnly: false,
					shared:   true,
				},
				{
					readOnly: false,
					shared:   false,
				},
			}

			for _, test := range testCases {
				err := a.Mount("/", "/", nil, test.readOnly, test.shared)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})

var _ = Describe("AlluxioFileUtils.IsMounted", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, err := a.IsMounted("/hbase")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "https://mirrors.bit.edu.cn/apache/hbase/stable  on  /hbase (web, capacity=-1B, used=-1B, read-only, not shared, properties={}) \n /underFSStorage  on  /  (local, capacity=0B, used=0B, not read-only, not shared, properties={})", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			testCases := []struct {
				alluxioPath    string
				expectedResult bool
			}{
				{
					alluxioPath:    "/spark",
					expectedResult: false,
				},
				{
					alluxioPath:    "/hbase",
					expectedResult: true,
				},
			}

			for _, test := range testCases {
				mounted, err := a.IsMounted(test.alluxioPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(mounted).To(Equal(test.expectedResult))
			}
		})
	})
})

var _ = Describe("AlluxioFileUtils.FindUnmountedAlluxioPaths", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec executes successfully", func() {
		const returnMessage = `s3://bucket/path/train on /cache (s3, capacity=-1B, used=-1B, not read-only, not shared, properties={alluxio.underfs.s3.inherit.acl=false, alluxio.underfs.s3.endpoint=s3endpoint, aws.secretKey=, aws.accessKeyId=})
/underFSStorage on / (local, capacity=0B, used=0B, not read-only, not shared, properties={})`

		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return returnMessage, "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should work correctly", func() {
			testCases := []struct {
				alluxioPaths           []string
				expectedUnmountedPaths []string
			}{
				{
					alluxioPaths:           []string{"/cache"},
					expectedUnmountedPaths: []string{},
				},
				{
					alluxioPaths:           []string{"/cache", "/cache2"},
					expectedUnmountedPaths: []string{"/cache2"},
				},
				{
					alluxioPaths:           []string{},
					expectedUnmountedPaths: []string{},
				},
				{
					alluxioPaths:           []string{"/cache2"},
					expectedUnmountedPaths: []string{"/cache2"},
				},
			}

			for _, test := range testCases {
				unmountedPaths, err := a.FindUnmountedAlluxioPaths(test.alluxioPaths)
				Expect(err).NotTo(HaveOccurred())
				if len(unmountedPaths) != 0 || len(test.expectedUnmountedPaths) != 0 {
					Expect(unmountedPaths).To(Equal(test.expectedUnmountedPaths))
				}
			}
		})
	})
})

var _ = Describe("AlluxioFileUtils.Ready", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return false", func() {
			ready := a.Ready()
			Expect(ready).To(BeFalse())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary: ", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should return true", func() {
			ready := a.Ready()
			Expect(ready).To(BeTrue())
		})
	})
})

var _ = Describe("AlluxioFileUtils.Du2", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, _, _, err := a.Du("/hbase")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "File Size     In Alluxio       Path\n577575561     0 (0%)           /hbase", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			ufs, cached, cachedPercentage, err := a.Du("/hbase")
			Expect(err).NotTo(HaveOccurred())
			Expect(ufs).To(Equal(int64(577575561)))
			Expect(cached).To(Equal(int64(0)))
			Expect(cachedPercentage).To(Equal("0%"))
		})
	})
})

var _ = Describe("AlluxioFileUtils.Count", func() {
	var (
		patch *gomonkey.Patches
	)

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when kubeclient exec works correctly", func() {
		BeforeEach(func() {
			out1, out2, out3 := 111, 222, 333
			mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
				if strings.Contains(p4[3], EXEC_ERR) {
					return "does not exist", "", errors.New("exec-error")
				} else if strings.Contains(p4[3], NEGATIVE_RES) {
					return "12324\t45463\t-9223372036854775808", "", nil
				} else if strings.Contains(p4[3], TOO_MANY_LINES) {
					return "1\n2\n3\n4\n", "1\n2\n3\n4\n", nil
				} else if strings.Contains(p4[3], DATA_NUM) {
					return "1\n2\t3", "1\n2\t3", nil
				} else if strings.Contains(p4[3], PARSE_ERR) {
					return "1\n1\tdududu\tbbb\t", "1\n1\t2\tbbb\t", nil
				} else {
					return fmt.Sprintf("first line!\n%d\t%d\t%d", out1, out2, out3), "", nil
				}
			}

			patch = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)
		})

		It("should return error", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(EXEC_ERR)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when negative res", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(NEGATIVE_RES)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when too many lines", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(TOO_MANY_LINES)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when data num not match", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(DATA_NUM)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when parse failed", func() {
			_, _, _, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(PARSE_ERR)
			Expect(err).To(HaveOccurred())
		})

		It("should not return error", func() {
			o1, o2, o3, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(FINE)
			Expect(err).NotTo(HaveOccurred())
			Expect(o1).To(Equal(int64(111)))
			Expect(o2).To(Equal(int64(222)))
			Expect(o3).To(Equal(int64(333)))
		})
	})
})

var _ = Describe("AlluxioFileUtils.GetFileCount", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecWithoutTimeoutErr)
		})

		It("should return an error", func() {
			_, err := a.GetFileCount()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Type: COUNTER, Value: 6,367,897", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecWithoutTimeoutCommon)
		})

		It("should not return an error", func() {
			fileCount, err := a.GetFileCount()
			Expect(err).NotTo(HaveOccurred())
			Expect(fileCount).To(Equal(int64(6367897)))
		})
	})
})

var _ = Describe("AlluxioFileUtils.ReportMetrics", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, err := a.ReportMetrics()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			_, err := a.ReportMetrics()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.ReportCapacity", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, err := a.ReportCapacity()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			_, err := a.ReportCapacity()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.exec", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecWithoutTimeoutErr)
		})

		It("should return an error", func() {
			_, _, err := a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecWithoutTimeoutCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Type: COUNTER, Value: 6,367,897", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecWithoutTimeoutCommon)
		})

		It("should not return an error", func() {
			_, _, err := a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, true)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("AlluxioFileUtils.MasterPodName", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			_, err := a.MasterPodName()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary: \n    Master Address: 192.168.0.193:20009\n    Web Port: 20010", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			address, err := a.MasterPodName()
			Expect(err).NotTo(HaveOccurred())
			Expect(address).To(Equal("192.168.0.193"))
		})
	})
})

var _ = Describe("AlluxioFileUtils.ExecMountScripts", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		BeforeEach(func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)
		})

		It("should return an error", func() {
			err := a.ExecMountScripts()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		BeforeEach(func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "test", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
		})

		It("should not return an error", func() {
			err := a.ExecMountScripts()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
