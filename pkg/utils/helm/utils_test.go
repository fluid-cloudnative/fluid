/*
Copyright 2023 The Fluid Authors.

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

package helm

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}

var _ = Describe("Helm", func() {
	Describe("InstallRelease", func() {
		var (
			lookPathPatch      *gomonkey.Patches
			statPatch          *gomonkey.Patches
			combineOutputPatch *gomonkey.Patches
		)

		AfterEach(func() {
			if lookPathPatch != nil {
				lookPathPatch.Reset()
			}
			if statPatch != nil {
				statPatch.Reset()
			}
			if combineOutputPatch != nil {
				combineOutputPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				err := InstallRelease("fluid", "default", "testValueFile", "testChartName")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when chart file does not exist", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				statPatch = gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, errors.New("fail to run the command")
				})

				err := InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when CombinedOutput fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				statPatch = gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				})
				combineOutputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "CombinedOutput", func(cmd *exec.Cmd) ([]byte, error) {
					return nil, errors.New("fail to run the command")
				})

				err := InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				statPatch = gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				})

				badValue := "test$bad"
				err := InstallRelease("fluid", badValue, "testValueFile", "/chart/fluid")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all conditions are met", func() {
			It("should install release successfully", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				statPatch = gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				})
				combineOutputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "CombinedOutput", func(cmd *exec.Cmd) ([]byte, error) {
					return []byte("test-output"), nil
				})

				err := InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("CheckRelease", func() {
		var (
			lookupPatch *gomonkey.Patches
			startPatch  *gomonkey.Patches
			waitPatch   *gomonkey.Patches
		)

		AfterEach(func() {
			if lookupPatch != nil {
				lookupPatch.Reset()
			}
			if startPatch != nil {
				startPatch.Reset()
			}
			if waitPatch != nil {
				waitPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookupPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				_, err := CheckRelease("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Start fails", func() {
			It("should return an error", func() {
				lookupPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				startPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Start", func(cmd *exec.Cmd) error {
					return errors.New("fail to run the command")
				})

				_, err := CheckRelease("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookupPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})

				badValue := "test$bad"
				_, err := CheckRelease("fluid", badValue)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Wait fails", func() {
			It("should return an error", func() {
				lookupPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				startPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Start", func(cmd *exec.Cmd) error {
					return nil
				})
				waitPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Wait", func(cmd *exec.Cmd) error {
					return errors.New("fail to run the command")
				})

				_, err := CheckRelease("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteRelease", func() {
		var (
			lookPathPatch *gomonkey.Patches
			outputPatch   *gomonkey.Patches
		)

		AfterEach(func() {
			if lookPathPatch != nil {
				lookPathPatch.Reset()
			}
			if outputPatch != nil {
				outputPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				err := DeleteRelease("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Output fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return nil, errors.New("fail to run the command")
				})

				err := DeleteRelease("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})

				badValue := "test$bad"
				err := DeleteRelease("fluid", badValue)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all conditions are met", func() {
			It("should delete release successfully", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return []byte("fluid:v0.6.0"), nil
				})

				err := DeleteRelease("fluid", "default")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("ListReleases", func() {
		var (
			lookPathPatch *gomonkey.Patches
			outputPatch   *gomonkey.Patches
		)

		AfterEach(func() {
			if lookPathPatch != nil {
				lookPathPatch.Reset()
			}
			if outputPatch != nil {
				outputPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				_, err := ListReleases("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Output fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return nil, errors.New("fail to run the command")
				})

				_, err := ListReleases("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})

				_, err := ListReleases("def$ault")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all conditions are met", func() {
			It("should list releases successfully", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return []byte("fluid:v0.6.0\nfluid:v0.5.0"), nil
				})

				release, err := ListReleases("default")
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(HaveLen(2))
			})
		})
	})

	Describe("ListReleaseMap", func() {
		var (
			lookPathPatch *gomonkey.Patches
			outputPatch   *gomonkey.Patches
		)

		AfterEach(func() {
			if lookPathPatch != nil {
				lookPathPatch.Reset()
			}
			if outputPatch != nil {
				outputPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				_, err := ListReleaseMap("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Output fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return nil, errors.New("fail to run the command")
				})

				_, err := ListReleaseMap("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})

				_, err := ListReleaseMap("def$ault")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all conditions are met", func() {
			It("should list release map successfully", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return []byte("fluid v0.6.0\nspark v0.5.0"), nil
				})

				release, err := ListReleaseMap("default")
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(HaveLen(2))
			})
		})
	})

	Describe("ListAllReleasesWithDetail", func() {
		var (
			lookPathPatch *gomonkey.Patches
			outputPatch   *gomonkey.Patches
		)

		AfterEach(func() {
			if lookPathPatch != nil {
				lookPathPatch.Reset()
			}
			if outputPatch != nil {
				outputPatch.Reset()
			}
		})

		Context("when LookPath fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "", errors.New("fail to run the command")
				})

				_, err := ListAllReleasesWithDetail("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Output fails", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return nil, errors.New("fail to run the command")
				})

				_, err := ListAllReleasesWithDetail("default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when namespace contains invalid characters", func() {
			It("should return an error", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})

				_, err := ListAllReleasesWithDetail("def$ault")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all conditions are met", func() {
			It("should list all releases with detail successfully", func() {
				lookPathPatch = gomonkey.ApplyFunc(exec.LookPath, func(file string) (string, error) {
					return "test-path", nil
				})
				outputPatch = gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", func(cmd *exec.Cmd) ([]byte, error) {
					return []byte("fluid default 1 2021-07-19 16:20:16.166658248 +0800 CST deployed fluid-0.6.0 0.6.0-3c06c0e\nspark default 2 2021-07-19 16:20:16.166658248 +0800 CST deployed spark-0.3.0 0.3.0-3c06c0e"), nil
				})

				release, err := ListAllReleasesWithDetail("default")
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(HaveLen(2))
			})
		})
	})

	Describe("DeleteReleaseIfExists", func() {
		var patches *gomonkey.Patches

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when CheckRelease fails", func() {
			It("should return an error", func() {
				patches = gomonkey.ApplyFunc(CheckRelease, func(name, namespace string) (exist bool, err error) {
					return false, errors.New("fail to run the command")
				})

				err := DeleteReleaseIfExists("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when release does not exist", func() {
			It("should not return an error", func() {
				patches = gomonkey.ApplyFunc(CheckRelease, func(name, namespace string) (exist bool, err error) {
					return false, nil
				})

				err := DeleteReleaseIfExists("fluid", "default")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when release exists but DeleteRelease fails", func() {
			It("should return an error", func() {
				patches = gomonkey.ApplyFunc(CheckRelease, func(name, namespace string) (exist bool, err error) {
					return true, nil
				})
				patches.ApplyFunc(DeleteRelease, func(name, namespace string) (err error) {
					return errors.New("fail to run the command")
				})

				err := DeleteReleaseIfExists("fluid", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when release exists and can be deleted", func() {
			It("should delete successfully", func() {
				patches = gomonkey.ApplyFunc(CheckRelease, func(name, namespace string) (exist bool, err error) {
					return true, nil
				})
				patches.ApplyFunc(DeleteRelease, func(name, namespace string) (err error) {
					return nil
				})

				err := DeleteReleaseIfExists("fluid", "default")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
