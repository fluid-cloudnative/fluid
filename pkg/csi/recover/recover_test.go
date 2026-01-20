/*
Copyright 2022 The Fluid Authors.

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

package recover

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testfuseRecoverPeriod = 30

var _ = Describe("FuseRecover", func() {
	Describe("initializeKubeletClient", func() {
		Context("when all environment variables are set correctly", func() {
			It("should initialize kubelet client successfully", func() {
				const (
					fakeToken          = "fakeToken"
					fakeNodeIP         = "fakeNodeIP"
					fakeClientCert     = ""
					fakeClientKey      = ""
					fakeKubeletTimeout = "120"
				)
				patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
					return []byte(fakeToken), nil
				})
				defer patch1.Reset()

				os.Setenv("NODE_IP", fakeNodeIP)
				os.Setenv("KUBELET_CLIENT_CERT", fakeClientCert)
				os.Setenv("KUBELET_CLIENT_KEY", fakeClientKey)
				os.Setenv("KUBELET_TIMEOUT", fakeKubeletTimeout)

				kubeletClient, err := initializeKubeletClient()
				Expect(err).NotTo(HaveOccurred())
				Expect(kubeletClient).NotTo(BeNil())
			})
		})

		Context("when KUBELET_TIMEOUT is not set", func() {
			It("should use default timeout", func() {
				patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
					return []byte("token"), nil
				})
				defer patch1.Reset()

				os.Setenv("NODE_IP", "192.168.1.1")
				os.Setenv("KUBELET_CLIENT_CERT", "")
				os.Setenv("KUBELET_CLIENT_KEY", "")
				os.Unsetenv("KUBELET_TIMEOUT")

				kubeletClient, err := initializeKubeletClient()
				Expect(err).NotTo(HaveOccurred())
				Expect(kubeletClient).NotTo(BeNil())
			})
		})

		Context("when token file cannot be read", func() {
			It("should return error", func() {
				patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
					return nil, fmt.Errorf("file not found")
				})
				defer patch1.Reset()

				kubeletClient, err := initializeKubeletClient()
				Expect(err).To(HaveOccurred())
				Expect(kubeletClient).To(BeNil())
			})
		})

		Context("when KUBELET_TIMEOUT is invalid", func() {
			It("should return error", func() {
				patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
					return []byte("token"), nil
				})
				defer patch1.Reset()

				os.Setenv("NODE_IP", "192.168.1.1")
				os.Setenv("KUBELET_TIMEOUT", "invalid")

				kubeletClient, err := initializeKubeletClient()
				Expect(err).To(HaveOccurred())
				Expect(kubeletClient).To(BeNil())
			})
		})
	})

	Describe("NewFuseRecover", func() {
		Context("when all parameters are valid", func() {
			It("should create new FuseRecover successfully", func() {
				fakeClient := fake.NewFakeClient()
				fakeRecorder := record.NewFakeRecorder(1)
				volumeLocks := utils.NewVolumeLocks()

				os.Setenv(utils.MountRoot, "/runtime-mnt")
				os.Setenv(FuseRecoveryPeriod, "5s")

				got, err := NewFuseRecover(fakeClient, fakeRecorder, fakeClient, volumeLocks)

				Expect(err).NotTo(HaveOccurred())
				Expect(got).NotTo(BeNil())
				Expect(got.KubeClient).To(Equal(fakeClient))
				Expect(got.ApiReader).To(Equal(fakeClient))
				Expect(got.Recorder).To(Equal(fakeRecorder))
				Expect(got.locks).To(Equal(volumeLocks))
				Expect(got.recoverFusePeriod).To(Equal(defaultFuseRecoveryPeriod))
				Expect(got.recoverWarningThreshold).To(Equal(defaultRecoverWarningThreshold))
			})
		})

		Context("when MountRoot environment variable is not set", func() {
			It("should return error", func() {
				fakeClient := fake.NewFakeClient()
				fakeRecorder := record.NewFakeRecorder(1)
				volumeLocks := utils.NewVolumeLocks()

				os.Unsetenv(utils.MountRoot)

				got, err := NewFuseRecover(fakeClient, fakeRecorder, fakeClient, volumeLocks)

				Expect(err).To(HaveOccurred())
				Expect(got).To(BeNil())
			})
		})

		Context("when RecoverWarningThreshold is set", func() {
			It("should use custom warning threshold", func() {
				fakeClient := fake.NewFakeClient()
				fakeRecorder := record.NewFakeRecorder(1)
				volumeLocks := utils.NewVolumeLocks()

				os.Setenv(utils.MountRoot, "/runtime-mnt")
				os.Setenv(RecoverWarningThreshold, "100")

				got, err := NewFuseRecover(fakeClient, fakeRecorder, fakeClient, volumeLocks)

				Expect(err).NotTo(HaveOccurred())
				Expect(got).NotTo(BeNil())
				Expect(got.recoverWarningThreshold).To(Equal(100))
			})
		})
	})

	Describe("Start", func() {
		It("should start the recovery process", func() {
			s := apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = corev1.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s)

			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: &mount.FakeMounter{},
				},
				KubeClient:              fakeClient,
				ApiReader:               fakeClient,
				Recorder:                record.NewFakeRecorder(1),
				recoverFusePeriod:       100 * time.Millisecond,
				recoverWarningThreshold: 50,
				locks:                   utils.NewVolumeLocks(),
			}

			patch1 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{}, nil
			})
			defer patch1.Reset()

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			go func() {
				err := r.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			time.Sleep(150 * time.Millisecond)
		})
	})

	Describe("runOnce", func() {
		It("should recover broken mount points", func() {
			dataset := &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jfsdemo",
					Namespace: "default",
				},
			}

			s := apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = corev1.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s, dataset)

			mockedFsMounts := map[string]string{}

			sourcePath := "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"
			targetPath := "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount"

			fakeMounter := &mount.FakeMounter{}
			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				},
				KubeClient:              fakeClient,
				ApiReader:               fakeClient,
				Recorder:                record.NewFakeRecorder(1),
				recoverFusePeriod:       testfuseRecoverPeriod,
				recoverWarningThreshold: 50,
				locks:                   utils.NewVolumeLocks(),
			}

			patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Mount", func(_ *mount.FakeMounter, source string, target string, _ string, _ []string) error {
				mockedFsMounts[source] = target
				return nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{{
					SourcePath:            sourcePath,
					MountPath:             targetPath,
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}}, nil
			})
			defer patch2.Reset()

			patch3 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
				return "default", "jfsdemo", nil
			})
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(r, "shouldRecover", func(mountPath string) (bool, error) {
				return true, nil
			})
			defer patch4.Reset()

			r.runOnce()

			target, exists := mockedFsMounts[sourcePath]
			Expect(exists).To(BeTrue())
			Expect(target).To(Equal(targetPath))
		})
	})

	Describe("recover", func() {
		Context("when GetBrokenMountPoints returns error", func() {
			It("should handle error gracefully", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fake.NewFakeClient(),
					ApiReader:               fake.NewFakeClient(),
					Recorder:                record.NewFakeRecorder(1),
					recoverWarningThreshold: 50,
					locks:                   utils.NewVolumeLocks(),
				}

				patch1 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
					return nil, fmt.Errorf("failed to get mount points")
				})
				defer patch1.Reset()

				Expect(func() { r.recover() }).NotTo(Panic())
			})
		})

		Context("when there are no broken mount points", func() {
			It("should complete without error", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fake.NewFakeClient(),
					ApiReader:               fake.NewFakeClient(),
					Recorder:                record.NewFakeRecorder(1),
					recoverWarningThreshold: 50,
					locks:                   utils.NewVolumeLocks(),
				}

				patch1 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
					return []mountinfo.MountPoint{}, nil
				})
				defer patch1.Reset()

				Expect(func() { r.recover() }).NotTo(Panic())
			})
		})
	})

	Describe("recoverBrokenMount", func() {
		Context("when mounting with default options", func() {
			It("should recover broken mount without error", func() {
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: &mount.FakeMounter{},
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "test",
				}
				err := r.recoverBrokenMount(point)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when mounting with readonly option", func() {
			It("should include ro option", func() {
				fakeMounter := &mount.FakeMounter{}
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test-source",
					MountPath:             "/test-target",
					FilesystemType:        "test",
					ReadOnly:              true,
					Count:                 1,
					NamespacedDatasetName: "test",
				}

				mountCalled := false
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Mount", func(_ *mount.FakeMounter, source string, target string, fstype string, options []string) error {
					mountCalled = true
					Expect(options).To(ContainElement("ro"))
					Expect(options).To(ContainElement("bind"))
					return nil
				})
				defer patch1.Reset()

				err := r.recoverBrokenMount(point)
				Expect(err).NotTo(HaveOccurred())
				Expect(mountCalled).To(BeTrue())
			})
		})

		Context("when mount fails", func() {
			It("should return without error but log the failure", func() {
				fakeMounter := &mount.FakeMounter{}
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "test",
				}

				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Mount", func(_ *mount.FakeMounter, source string, target string, fstype string, options []string) error {
					return fmt.Errorf("mount failed")
				})
				defer patch1.Reset()

				err := r.recoverBrokenMount(point)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("umountDuplicate", func() {
		Context("when count is greater than 1", func() {
			It("should unmount duplicate mount points", func() {
				fakeMounter := &mount.FakeMounter{}
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 3,
					NamespacedDatasetName: "test",
				}

				unmountCount := 0
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Unmount", func(_ *mount.FakeMounter, target string) error {
					unmountCount++
					return nil
				})
				defer patch1.Reset()

				r.umountDuplicate(point)
				Expect(unmountCount).To(Equal(2))
			})
		})

		Context("when unmount fails", func() {
			It("should not panic", func() {
				fakeMounter := &mount.FakeMounter{}
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 3,
					NamespacedDatasetName: "test",
				}

				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Unmount", func(_ *mount.FakeMounter, target string) error {
					return fmt.Errorf("unmount failed")
				})
				defer patch1.Reset()

				Expect(func() { r.umountDuplicate(point) }).NotTo(Panic())
			})
		})

		Context("when count is 1", func() {
			It("should not unmount anything", func() {
				fakeMounter := &mount.FakeMounter{}
				r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				}}
				point := mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "test",
				}

				unmountCount := 0
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Unmount", func(_ *mount.FakeMounter, target string) error {
					unmountCount++
					return nil
				})
				defer patch1.Reset()

				r.umountDuplicate(point)
				Expect(unmountCount).To(Equal(0))
			})
		})
	})

	Describe("eventRecord", func() {
		var (
			s          *apimachineryRuntime.Scheme
			dataset    *v1alpha1.Dataset
			pv         *corev1.PersistentVolume
			fakeClient client.Client
		)

		BeforeEach(func() {
			s = apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = scheme.AddToScheme(s)

			dataset = &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jfsdemo",
					Namespace: "default",
				},
			}
			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default-jfsdemo",
				},
			}
		})

		Context("when recording FuseRecoverSucceed event", func() {
			It("should record event without panic", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed) }).NotTo(Panic())
			})
		})

		Context("when recording FuseRecoverFailed event", func() {
			It("should record event without panic", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeWarning, common.FuseRecoverFailed) }).NotTo(Panic())
			})
		})

		Context("when recording FuseUmountDuplicate event", func() {
			It("should record event without panic", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 100,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeWarning, common.FuseUmountDuplicate) }).NotTo(Panic())
			})
		})

		Context("when namespacedName has less than 2 parts", func() {
			It("should handle gracefully", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "invalid",
				}

				Expect(func() { r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed) }).NotTo(Panic())
			})
		})

		Context("when GetNamespacedNameByVolumeId returns error", func() {
			It("should handle error gracefully", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "", "", fmt.Errorf("volume not found")
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed) }).NotTo(Panic())
			})
		})

		Context("when dataset is not found", func() {
			It("should handle error gracefully", func() {
				fakeClient = fake.NewFakeClientWithScheme(s, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed) }).NotTo(Panic())
			})
		})

		Context("when PV name is mismatched", func() {
			It("should record event without panic", func() {
				pv := &corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name: "other-pv",
					},
				}
				fakeClient = fake.NewFakeClientWithScheme(s, dataset, pv)
				r := &FuseRecover{
					KubeClient: fakeClient,
					ApiReader:  fakeClient,
					Recorder:   record.NewFakeRecorder(1),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "jfsdemo",
				}

				patch1 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch1.Reset()

				Expect(func() { r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed) }).NotTo(Panic())
			})
		})
	})

	Describe("shouldRecover", func() {
		Context("when mount point does not exist", func() {
			It("should return false without error", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
				}

				fakeMounter := mount.New("")
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "IsLikelyNotMountPoint", func(_ *mount.Mounter, file string) (bool, error) {
					return true, os.ErrNotExist
				})
				defer patch1.Reset()

				should, err := r.shouldRecover("/nonexistent/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when mount point is not mounted", func() {
			It("should return false without error", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
				}

				fakeMounter := mount.New("")
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "IsLikelyNotMountPoint", func(_ *mount.Mounter, file string) (bool, error) {
					return true, nil
				})
				defer patch1.Reset()

				should, err := r.shouldRecover("/some/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when unexpected error occurs", func() {
			It("should return error", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
				}

				fakeMounter := mount.New("")
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "IsLikelyNotMountPoint", func(_ *mount.Mounter, file string) (bool, error) {
					return false, fmt.Errorf("unexpected error")
				})
				defer patch1.Reset()

				should, err := r.shouldRecover("/error/path")
				Expect(err).To(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when mount point is valid", func() {
			It("should return true without error", func() {
				r := &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
				}

				fakeMounter := mount.New("")
				patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "IsLikelyNotMountPoint", func(_ *mount.Mounter, file string) (bool, error) {
					return false, nil
				})
				defer patch1.Reset()

				should, err := r.shouldRecover("/valid/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeTrue())
			})
		})
	})

	Describe("doRecover", func() {
		var (
			s          *apimachineryRuntime.Scheme
			dataset    *v1alpha1.Dataset
			fakeClient client.Client
			r          *FuseRecover
		)

		BeforeEach(func() {
			s = apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = corev1.AddToScheme(s)

			dataset = &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jfsdemo",
					Namespace: "default",
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(s, dataset)
		})

		Context("when lock cannot be acquired", func() {
			It("should skip recovery", func() {
				locks := utils.NewVolumeLocks()
				locks.TryAcquire("/test/path")

				r = &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fakeClient,
					ApiReader:               fakeClient,
					Recorder:                record.NewFakeRecorder(10),
					recoverWarningThreshold: 50,
					locks:                   locks,
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/test/source",
					MountPath:             "/test/path",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "default-jfsdemo",
				}

				Expect(func() { r.doRecover(point) }).NotTo(Panic())
			})
		})

		Context("when shouldRecover returns error", func() {
			It("should skip recovery", func() {
				r = &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fakeClient,
					ApiReader:               fakeClient,
					Recorder:                record.NewFakeRecorder(10),
					recoverWarningThreshold: 50,
					locks:                   utils.NewVolumeLocks(),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/test/source",
					MountPath:             "/test/path",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyPrivateMethod(r, "shouldRecover", func(mountPath string) (bool, error) {
					return false, fmt.Errorf("check failed")
				})
				defer patch1.Reset()

				Expect(func() { r.doRecover(point) }).NotTo(Panic())
			})
		})

		Context("when shouldRecover returns false", func() {
			It("should skip recovery", func() {
				r = &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fakeClient,
					ApiReader:               fakeClient,
					Recorder:                record.NewFakeRecorder(10),
					recoverWarningThreshold: 50,
					locks:                   utils.NewVolumeLocks(),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/test/source",
					MountPath:             "/test/path",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyPrivateMethod(r, "shouldRecover", func(mountPath string) (bool, error) {
					return false, nil
				})
				defer patch1.Reset()

				Expect(func() { r.doRecover(point) }).NotTo(Panic())
			})
		})

		Context("when count exceeds warning threshold", func() {
			It("should umount duplicates and record warning event", func() {
				r = &FuseRecover{
					SafeFormatAndMount: mount.SafeFormatAndMount{
						Interface: &mount.FakeMounter{},
					},
					KubeClient:              fakeClient,
					ApiReader:               fakeClient,
					Recorder:                record.NewFakeRecorder(10),
					recoverWarningThreshold: 50,
					locks:                   utils.NewVolumeLocks(),
				}

				point := mountinfo.MountPoint{
					SourcePath:            "/test/source",
					MountPath:             "/test/path",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 100,
					NamespacedDatasetName: "default-jfsdemo",
				}

				patch1 := ApplyPrivateMethod(r, "shouldRecover", func(mountPath string) (bool, error) {
					return true, nil
				})
				defer patch1.Reset()

				patch2 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
					return "default", "jfsdemo", nil
				})
				defer patch2.Reset()

				unmountCalled := false
				patch3 := ApplyPrivateMethod(r, "umountDuplicate", func(point mountinfo.MountPoint) {
					unmountCalled = true
				})
				defer patch3.Reset()

				r.doRecover(point)
				Expect(unmountCalled).To(BeTrue())
			})
		})
	})
})
