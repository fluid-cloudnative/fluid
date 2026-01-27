package discovery

import (
	"errors"
	nativeLog "log"
	"sync"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("ResourceEnabled", func() {
	BeforeEach(func() {
		globalDiscovery = map[string]bool{
			"dataload":       true,
			"alluxioruntime": true,
			"dataset":        true,
			"disabled":       false,
		}
	})

	Context("when checking enabled resources", func() {
		It("should return true for dataset", func() {
			got := globalDiscovery.ResourceEnabled("dataset")
			Expect(got).To(BeTrue())
		})

		It("should return true for alluxioruntime", func() {
			got := globalDiscovery.ResourceEnabled("alluxioruntime")
			Expect(got).To(BeTrue())
		})

		It("should return false for databackup (not exists)", func() {
			got := globalDiscovery.ResourceEnabled("databackup")
			Expect(got).To(BeFalse())
		})

		It("should return false for disabled resource (exists but false)", func() {
			got := globalDiscovery.ResourceEnabled("disabled")
			Expect(got).To(BeFalse())
		})
	})
})

var _ = Describe("discoverFluidResourcesInCluster", func() {
	BeforeEach(func() {
		globalDiscovery = map[string]bool{}
	})

	Context("when discovering resources successfully", func() {
		It("should discover multiple resources", func() {
			patchedFunc := func(groupVersion string) (*metav1.APIResourceList, error) {
				return &metav1.APIResourceList{
					APIResources: []metav1.APIResource{
						{SingularName: "dataset"},
						{SingularName: "alluxioruntime"},
						{SingularName: "dataload"},
					},
				}, nil
			}

			patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
				return &rest.Config{}
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
				return &discovery.DiscoveryClient{}
			})
			defer patch2.Reset()

			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", patchedFunc)
			defer patch3.Reset()

			discoverFluidResourcesInCluster()

			wantResources := fluidDiscovery(map[string]bool{
				"dataset":        true,
				"alluxioruntime": true,
				"dataload":       true,
			})
			Expect(globalDiscovery).To(Equal(wantResources))
		})

		It("should handle uppercase resource names", func() {
			patchedFunc := func(groupVersion string) (*metav1.APIResourceList, error) {
				return &metav1.APIResourceList{
					APIResources: []metav1.APIResource{
						{SingularName: "Dataset"},
						{SingularName: "AlluxioRuntime"},
					},
				}, nil
			}

			patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
				return &rest.Config{}
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
				return &discovery.DiscoveryClient{}
			})
			defer patch2.Reset()

			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", patchedFunc)
			defer patch3.Reset()

			discoverFluidResourcesInCluster()

			wantResources := fluidDiscovery(map[string]bool{
				"dataset":        true,
				"alluxioruntime": true,
			})
			Expect(globalDiscovery).To(Equal(wantResources))
		})
	})

	Context("when handling error cases", func() {

		It("should panic on discovery client error with retry exhaustion", func() {
			callCount := 0
			patchedFunc := func(groupVersion string) (*metav1.APIResourceList, error) {
				callCount++
				return nil, errors.New("connection refused")
			}

			patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
				return &rest.Config{}
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
				return &discovery.DiscoveryClient{}
			})
			defer patch2.Reset()

			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", patchedFunc)
			defer patch3.Reset()

			// Patch time.Sleep to skip delays during retries
			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {
				// No-op: skip actual sleep
			})
			defer patchSleep.Reset()

			fatalCalled := false
			patchFatal := gomonkey.ApplyFunc(nativeLog.Fatalf, func(format string, v ...interface{}) {
				fatalCalled = true
				panic("log.Fatalf called")
			})
			defer patchFatal.Reset()

			// Use channel-based test with timeout
			done := make(chan bool, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- true
					}
				}()
				discoverFluidResourcesInCluster()
				done <- false
			}()

			Eventually(done, 5*time.Second).Should(Receive(BeTrue()))
			Expect(fatalCalled).To(BeTrue())
		})

		It("should panic when no CRDs are found", func() {
			patchedFunc := func(groupVersion string) (*metav1.APIResourceList, error) {
				return &metav1.APIResourceList{
					APIResources: []metav1.APIResource{},
				}, nil
			}

			patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
				return &rest.Config{}
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
				return &discovery.DiscoveryClient{}
			})
			defer patch2.Reset()

			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", patchedFunc)
			defer patch3.Reset()

			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {
				// No-op: skip actual sleep
			})
			defer patchSleep.Reset()

			fatalCalled := false
			patchFatal := gomonkey.ApplyFunc(nativeLog.Fatalf, func(format string, v ...interface{}) {
				fatalCalled = true
				panic("log.Fatalf called")
			})
			defer patchFatal.Reset()

			done := make(chan bool, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- true
					}
				}()
				discoverFluidResourcesInCluster()
				done <- false
			}()

			Eventually(done, 5*time.Second).Should(Receive(BeTrue()))
			Expect(fatalCalled).To(BeTrue())
		})
	})
})

var _ = Describe("InitDiscovery", func() {
	BeforeEach(func() {
		globalDiscovery = map[string]bool{}
	})

	It("should initialize discovery with resources", func() {
		patchedFunc := func(groupVersion string) (*metav1.APIResourceList, error) {
			return &metav1.APIResourceList{
				APIResources: []metav1.APIResource{
					{SingularName: "dataset"},
					{SingularName: "juicefsruntime"},
				},
			}, nil
		}

		patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
			return &rest.Config{}
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
			return &discovery.DiscoveryClient{}
		})
		defer patch2.Reset()

		var fakeClient *discovery.DiscoveryClient
		patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", patchedFunc)
		defer patch3.Reset()

		initDiscovery()

		wantResources := fluidDiscovery(map[string]bool{
			"dataset":        true,
			"juicefsruntime": true,
		})
		Expect(globalDiscovery).To(Equal(wantResources))
	})
})

var _ = Describe("GetFluidDiscovery", func() {
	var (
		originalDiscovery fluidDiscovery
	)

	BeforeEach(func() {
		// Save original state

		originalDiscovery = globalDiscovery

		// Reset for this test - NOTE: This is a workaround since sync.Once can't be truly reset
		// In production code, consider using a factory pattern or dependency injection
		once = sync.Once{}
		globalDiscovery = nil
	})

	AfterEach(func() {
		// Restore original state

		globalDiscovery = originalDiscovery
	})

	It("should initialize discovery on first call", func() {
		want := fluidDiscovery(map[string]bool{
			"foo": true,
			"bar": true,
		})
		patch := gomonkey.ApplyFunc(initDiscovery, func() {
			globalDiscovery = want
		})
		defer patch.Reset()

		got := GetFluidDiscovery()
		Expect(got).To(Equal(want))
	})

	It("should use existing discovery without re-initialization", func() {
		// Set up pre-existing discovery state
		existingDiscovery := fluidDiscovery(map[string]bool{
			"existing": true,
		})
		globalDiscovery = existingDiscovery
		once = sync.Once{} // Fresh Once that hasn't been called

		// Mark once as already called by calling Do
		once.Do(func() {})

		// Verify initDiscovery is NOT called
		initCalled := false
		patch := gomonkey.ApplyFunc(initDiscovery, func() {
			initCalled = true
		})
		defer patch.Reset()

		got := GetFluidDiscovery()
		Expect(got).To(Equal(existingDiscovery))
		Expect(initCalled).To(BeFalse())
	})
})
