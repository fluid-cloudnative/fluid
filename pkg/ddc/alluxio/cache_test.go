/*
Copyright 2021 The Fluid Authors.

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

package alluxio

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Cache related tests", Label("pkg.ddc.alluxio.cache_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjects  mockedObjects
		client         client.Client
		resources      []runtime.Object
	)
	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjects = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjects.MasterSts,
			mockedObjects.WorkerSts,
			mockedObjects.FuseDs,
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.runtimeInfo.SetFuseName(engine.getFuseName())
		engine.Client = client
	})

	Describe("Test AlluxioEngine.queryCacheStatus()", func() {

		When("dataset's ufs total size is not empty", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = "16.16MiB"
			})

			Context("and cached size is 0B", func() {
				It("should successfully query cache status and cached percentage should be 0%", func() {
					patch1 := ApplyMethodFunc(engine, "GetReportSummary", func() (string, error) {
						summary := mockAlluxioReportSummary("0B", "19.07MB")
						return summary, nil
					})
					defer patch1.Reset()
					patch2 := ApplyMethodFunc(engine, "GetCacheHitStates", func() cacheHitStates {
						return cacheHitStates{
							bytesReadLocal:  12345678,
							bytesReadUfsAll: 87654321,
						}
					})
					defer patch2.Reset()

					cacheStates, err := engine.queryCacheStatus()
					Expect(err).To(BeNil())
					Expect(cacheStates.cached).To(Equal("0.00B"))
					Expect(cacheStates.cacheCapacity).To(Equal("19.07MiB"))
					Expect(cacheStates.cachedPercentage).To(Equal("0.0%"))
					Expect(cacheStates.cacheHitStates.bytesReadLocal).To(Equal(int64(12345678)))
					Expect(cacheStates.cacheHitStates.bytesReadUfsAll).To(Equal(int64(87654321)))
				})
			})

			Context("and cache size is half of the ufs total size", func() {
				It("should successfully query cache status and cached percentage should be 50%", func() {
					patch1 := ApplyMethodFunc(engine, "GetReportSummary", func() (string, error) {
						summary := mockAlluxioReportSummary("8.08MB", "19.07MB")
						return summary, nil
					})
					defer patch1.Reset()
					patch2 := ApplyMethodFunc(engine, "GetCacheHitStates", func() cacheHitStates {
						return cacheHitStates{
							bytesReadLocal:  12345678,
							bytesReadUfsAll: 87654321,
						}
					})
					defer patch2.Reset()

					cacheStates, err := engine.queryCacheStatus()
					Expect(err).To(BeNil())
					Expect(cacheStates.cached).To(Equal("8.08MiB"))
					Expect(cacheStates.cacheCapacity).To(Equal("19.07MiB"))
					Expect(cacheStates.cachedPercentage).To(Equal("50.0%"))
					Expect(cacheStates.cacheHitStates.bytesReadLocal).To(Equal(int64(12345678)))
					Expect(cacheStates.cacheHitStates.bytesReadUfsAll).To(Equal(int64(87654321)))
				})
			})
		})

		When("dataset's ufs total size is [Calculating]", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = metadataSyncNotDoneMsg
			})

			It("should successfully query cache status and cached percentage is empty", func() {
				patch1 := ApplyMethodFunc(engine, "GetReportSummary", func() (string, error) {
					summary := mockAlluxioReportSummary("0B", "19.07MB")
					return summary, nil
				})
				defer patch1.Reset()
				patch2 := ApplyMethodFunc(engine, "GetCacheHitStates", func() cacheHitStates {
					return cacheHitStates{
						bytesReadLocal:  12345678,
						bytesReadUfsAll: 87654321,
					}
				})
				defer patch2.Reset()

				cacheStates, err := engine.queryCacheStatus()
				Expect(err).To(BeNil())
				Expect(cacheStates.cachedPercentage).To(HaveLen(0))
			})
		})

		When("dataset's ufs total size is empty", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = ""
			})

			It("should successfully query cache status and cached percentage is empty", func() {
				patch1 := ApplyMethodFunc(engine, "GetReportSummary", func() (string, error) {
					summary := mockAlluxioReportSummary("0B", "19.07MB")
					return summary, nil
				})
				defer patch1.Reset()
				patch2 := ApplyMethodFunc(engine, "GetCacheHitStates", func() cacheHitStates {
					return cacheHitStates{
						bytesReadLocal:  12345678,
						bytesReadUfsAll: 87654321,
					}
				})
				defer patch2.Reset()

				cacheStates, err := engine.queryCacheStatus()
				Expect(err).To(BeNil())
				Expect(cacheStates.cachedPercentage).To(HaveLen(0))
			})
		})
	})

	Describe("Test AlluxioEngine.GetCacheHitStates()", func() {
		When("first time to call GetCacheHitStates()", func() {
			BeforeEach(func() {
				engine.lastCacheHitStates = nil
			})

			It("should parse the cache hit states and store it into lastCacheHitStates", func() {
				patch1 := gomonkey.ApplyMethodFunc(engine, "GetReportMetrics", func() (string, error) {
					return mockAlluxioReportMetrics(
						"16.00MB",
						"500KB/MIN",
						"30.00MB",
						"700KB/MIN",
						"20.00MB",
						"1.1MB/MIN"), nil
				})
				defer patch1.Reset()

				gotCacheHitStates := engine.GetCacheHitStates()

				Expect(gotCacheHitStates.bytesReadLocal).To(Equal(int64(16 << 20) /*16.00MB*/))
				Expect(gotCacheHitStates.bytesReadRemote).To(Equal(int64(30 << 20) /*30.00MB*/))
				Expect(gotCacheHitStates.bytesReadUfsAll).To(Equal(int64(20 << 20) /*20.00MB*/))
				Expect(engine.lastCacheHitStates).NotTo(BeNil())
				// Expect(cacheHitStates.bytesReadLocal).To(Equal())
			})
		})

		When("second time to call GetCacheHitStates", func() {
			BeforeEach(func() {
				engine.lastCacheHitStates = &cacheHitStates{
					bytesReadLocal:  16 << 20,
					bytesReadRemote: 30 << 20,
					bytesReadUfsAll: 20 << 20,
				}
			})
			Context("when interval between lastCacheHitStates and now is less than 1 minute", func() {
				BeforeEach(func() {
					engine.lastCacheHitStates.timestamp = time.Now().Add(-1 * time.Second)
				})
				It("should return lastCacheHitStates", func() {
					gotCacheHitStates := engine.GetCacheHitStates()
					Expect(gotCacheHitStates.bytesReadLocal).To(Equal(int64(16 << 20)))
					Expect(gotCacheHitStates.bytesReadRemote).To(Equal(int64(30 << 20)))
					Expect(gotCacheHitStates.bytesReadUfsAll).To(Equal(int64(20 << 20)))
				})
			})

			Context("when interval between lastCacheHitStates and now is greater than 1 minute", func() {
				BeforeEach(func() {
					engine.lastCacheHitStates.timestamp = time.Now().Add(-2 * time.Minute)
				})
				It("should calculate cache hit states", func() {
					patch1 := gomonkey.ApplyMethodFunc(engine, "GetReportMetrics", func() (string, error) {
						return mockAlluxioReportMetrics(
							"21.00MB", // 16.00MB + 5.00MB
							"2.00MB/MIN",
							"40.00MB", // 30.00MB + 10.00MB
							"1.00MB/MIN",
							"55.00MB", // 20.00MB + 35.00MB
							"1.00MB/MIN"), nil
					})
					defer patch1.Reset()

					gotCacheHitStates := engine.GetCacheHitStates()
					Expect(gotCacheHitStates.bytesReadLocal).To(Equal(int64(21 << 20)))
					Expect(gotCacheHitStates.bytesReadRemote).To(Equal(int64(40 << 20)))
					Expect(gotCacheHitStates.bytesReadUfsAll).To(Equal(int64(55 << 20)))

					Expect(gotCacheHitStates.cacheHitRatio).To(Equal("30.0%"))  // (5.00MB + 10.00MB) / (5.00MB + 10.00MB + 35.00MB)
					Expect(gotCacheHitStates.localHitRatio).To(Equal("10.0%"))  // 5.00MB / (5.00MB + 10.00MB + 35.00MB)
					Expect(gotCacheHitStates.remoteHitRatio).To(Equal("20.0%")) // 10.00MB / (5.00MB + 10.00MB + 35.00MB)

					Expect(gotCacheHitStates.localThroughputRatio).To(Equal("50.0%"))  // 2.00MB/min / (2.00MB/min + 1.00MB/MIN + 1.00MB/MIN)
					Expect(gotCacheHitStates.remoteThroughputRatio).To(Equal("25.0%")) // 1.00MB/min / (2.00MB/min + 1.00MB/MIN + 1.00MB/MIN)
					Expect(gotCacheHitStates.cacheThroughputRatio).To(Equal("75.0%"))  // (1.00MB/min + 2.00MB/min) / (2.00MB/min + 1.00MB/MIN + 1.00MB/MIN)
				})
			})
		})
	})
})

// TestPatchDatasetStatus verifies that the patchDatasetStatus method correctly calculates
// and updates the cached data percentage in a Dataset's status. It runs multiple test cases
// with different total and cached values, and checks whether the computed percentage matches
// the expected result.
func TestPatchDatasetStatus(t *testing.T) {
	engine := &AlluxioEngine{}
	testCases := []struct {
		total      string
		cached     string
		percentage string
	}{
		{
			total:      "100",
			cached:     "10",
			percentage: "10.0%",
		},
		{
			total:      "100",
			cached:     "50",
			percentage: "50.0%",
		},
	}
	for _, testCase := range testCases {
		dataset := &datav1alpha1.Dataset{
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: testCase.total,
			},
		}
		states := &cacheStates{
			cached: testCase.cached,
		}
		engine.patchDatasetStatus(dataset, states)
		if states.cachedPercentage != testCase.percentage {
			t.Errorf(" want %s, got %s", testCase.percentage, states.cachedPercentage)
		}
	}
}

// TestInvokeCleanCache tests the behavior of the invokeCleanCache function in the AlluxioEngine.
// It simulates different StatefulSet statuses using a fake client to verify whether the function
// correctly determines if an error should be returned based on the readiness state of the replicas.
//
// Parameters:
// - t (*testing.T): The testing context used to run assertions.
//
// Returns:
// - None. The function uses t.Errorf to report test failures.
func TestInvokeCleanCache(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}
	objs := []runtime.Object{}
	for _, masterInput := range masterInputs {
		objs = append(objs, masterInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	testCases := []struct {
		name      string
		namespace string
		isErr     bool
	}{
		{
			name:      "hadoop",
			namespace: "fluid",
			isErr:     false,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			isErr:     true,
		},
		{
			name:      "none",
			namespace: "fluid",
			isErr:     false,
		},
	}
	for _, testCase := range testCases {
		engine := &AlluxioEngine{
			Client:    fakeClient,
			namespace: testCase.namespace,
			name:      testCase.name,
			Log:       fake.NullLogger(),
		}
		err := engine.invokeCleanCache("")
		isErr := err != nil
		if isErr != testCase.isErr {
			t.Errorf("test-name:%s want %t, got %t", testCase.name, testCase.isErr, isErr)
		}
	}
}

// TestAlluxioEngine_getGracefulShutdownLimits tests the getGracefulShutdownLimits method of the AlluxioEngine.
// This test verifies the correct retrieval of graceful shutdown retry limits under different runtime configurations.
// It covers three scenarios:
// 1. When no CleanCachePolicy is specified, ensuring the default value (3) is returned.
// 2. When CleanCachePolicy with MaxRetryAttempts is set, ensuring the configured value is returned.
// 3. An error scenario where the runtime object might not be properly retrieved, testing error handling.
// The test uses a fake Kubernetes client to simulate runtime object states and validate the method's behavior.
//
// Parameters:
// - t: Testing handler for test reporting and control.
func TestAlluxioEngine_getGracefulShutdownLimits(t *testing.T) {
	type fields struct {
		runtime            *datav1alpha1.AlluxioRuntime
		name               string
		namespace          string
		runtimeType        string
		Log                logr.Logger
		Client             client.Client
		retryShutdown      int32
		initImage          string
		MetadataSyncDoneCh chan base.MetadataSyncResult
		UnitTest           bool
		Recorder           record.EventRecorder
	}
	tests := []struct {
		name                       string
		fields                     fields
		wantGracefulShutdownLimits int32
		wantErr                    bool
	}{
		// TODO: Add test cases.
		{
			name: "no_clean_cache_policy",
			fields: fields{
				name:      "noCleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noCleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				},
			},
			wantGracefulShutdownLimits: 3,
			wantErr:                    false,
		}, {
			name: "clean_cache_policy",
			fields: fields{
				name:      "cleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantGracefulShutdownLimits: 12,
			wantErr:                    false,
		}, {
			name: "test_err",
			fields: fields{
				name:      "err",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "notFoundErr",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantGracefulShutdownLimits: 0,
			wantErr:                    true,
		},
	}
	objs := []runtime.Object{}
	for _, tt := range tests {
		objs = append(objs, tt.fields.runtime.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				runtimeType:        tt.fields.runtimeType,
				Log:                tt.fields.Log,
				Client:             fakeClient,
				retryShutdown:      tt.fields.retryShutdown,
				initImage:          tt.fields.initImage,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
				Recorder:           tt.fields.Recorder,
			}
			gotGracefulShutdownLimits, err := e.getGracefulShutdownLimits()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.getGracefulShutdownLimits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGracefulShutdownLimits != tt.wantGracefulShutdownLimits {
				t.Errorf("AlluxioEngine.getGracefulShutdownLimits() = %v, want %v", gotGracefulShutdownLimits, tt.wantGracefulShutdownLimits)
			}
		})
	}
}

// TestAlluxioEngine_getCleanCacheGracePeriodSeconds verifies the behavior of
// the AlluxioEngine.getCleanCacheGracePeriodSeconds() method.
// It covers the following scenarios:
// 1. When no CleanCachePolicy is defined in the runtime spec, the method should return the default value (60 seconds).
// 2. When GracePeriodSeconds is explicitly set in CleanCachePolicy, the method should return the configured value.
// 3. When the runtime object is not found or the GracePeriodSeconds is not defined properly, it should return an error.
// The test uses a fake Kubernetes client initialized with deep-copied runtime objects to simulate the environment.
func TestAlluxioEngine_getCleanCacheGracePeriodSeconds(t *testing.T) {
	type fields struct {
		runtime       *datav1alpha1.AlluxioRuntime
		name          string
		namespace     string
		runtimeType   string
		Log           logr.Logger
		Client        client.Client
		retryShutdown int32
		initImage     string
	}
	tests := []struct {
		name                             string
		fields                           fields
		wantCleanCacheGracePeriodSeconds int32
		wantErr                          bool
	}{
		// TODO: Add test cases.
		{
			name: "no_clean_cache_policy",
			fields: fields{
				name:      "noCleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noCleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				},
			},
			wantCleanCacheGracePeriodSeconds: 60,
			wantErr:                          false,
		}, {
			name: "clean_cache_policy",
			fields: fields{
				name:      "cleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								GracePeriodSeconds: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantCleanCacheGracePeriodSeconds: 12,
			wantErr:                          false,
		}, {
			name: "test_err",
			fields: fields{
				name:      "notFoundError",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "notFound",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantCleanCacheGracePeriodSeconds: 0,
			wantErr:                          true,
		},
	}

	objs := []runtime.Object{}
	for _, tt := range tests {
		objs = append(objs, tt.fields.runtime.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:       tt.fields.runtime,
				name:          tt.fields.name,
				namespace:     tt.fields.namespace,
				runtimeType:   tt.fields.runtimeType,
				Log:           tt.fields.Log,
				Client:        fakeClient,
				retryShutdown: tt.fields.retryShutdown,
				initImage:     tt.fields.initImage,
			}
			gotCleanCacheGracePeriodSeconds, err := e.getCleanCacheGracePeriodSeconds()
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase %v AlluxioEngine.getCleanCacheGracePeriodSeconds() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotCleanCacheGracePeriodSeconds != tt.wantCleanCacheGracePeriodSeconds {
				t.Errorf("testcase %v AlluxioEngine.getCleanCacheGracePeriodSeconds() = %v, want %v", tt.name, gotCleanCacheGracePeriodSeconds, tt.wantCleanCacheGracePeriodSeconds)
			}
		})
	}
}
