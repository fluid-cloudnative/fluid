package kubeclient

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// testScheme is kept as package-level variable for backwards compatibility with other test files
var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = rbacv1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)
}

var _ = Describe("Pod Utilities", func() {
	Describe("GetPVCNamesFromPod", func() {
		It("should extract PVC names from pod volumes", func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			pod := corev1.Pod{}
			var pvcNamesWant []string
			for i := 1; i <= 30; i++ {
				switch r.Intn(4) {
				case 0:
					pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
						Name: "volume" + strconv.Itoa(i),
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/tmp/data" + strconv.Itoa(i),
							},
						},
					})
				case 1:
					pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
						Name: "volume" + strconv.Itoa(i),
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc" + strconv.Itoa(i),
								ReadOnly:  true,
							},
						},
					})
					pvcNamesWant = append(pvcNamesWant, "pvc"+strconv.Itoa(i))
				case 2:
					pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
						Name: "volume" + strconv.Itoa(i),
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					})
				case 3:
					pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
						Name: "volume" + strconv.Itoa(i),
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server:   "172.0.0." + strconv.Itoa(i),
								Path:     "/data" + strconv.Itoa(i),
								ReadOnly: true,
							},
						},
					})
				}
			}
			pvcNames := GetPVCNamesFromPod(&pod)
			Expect(pvcNames).To(Equal(pvcNamesWant))
		})
	})

	Describe("IsCompletePod", func() {
		var (
			namespace  string
			pods       []*corev1.Pod
			fakeClient client.Client
		)

		BeforeEach(func() {
			namespace = "default"
			pods = []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod3", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodFailed},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod4",
						Namespace:         namespace,
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
						Finalizers:        []string{"test.finalizer"},
					},
					Spec: corev1.PodSpec{},
				},
			}

			testPods := []runtime.Object{}
			for _, pod := range pods {
				testPods = append(testPods, pod.DeepCopy())
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testPods...)
		})

		type testCase struct {
			name      string
			namespace string
			want      bool
		}

		DescribeTable("should correctly identify complete pods",
			func(tc testCase) {
				var pod corev1.Pod
				var podToTest *corev1.Pod
				key := types.NamespacedName{
					Namespace: tc.namespace,
					Name:      tc.name,
				}
				_ = fakeClient.Get(context.TODO(), key, &pod)
				if len(pod.Name) == 0 {
					podToTest = nil
				} else {
					podToTest = &pod
				}
				Expect(IsCompletePod(podToTest)).To(Equal(tc.want))
			},
			Entry("Pod doesn't exist", testCase{name: "notExist", namespace: "default", want: false}),
			Entry("Pod is running", testCase{name: "pod1", namespace: "default", want: false}),
			Entry("Pod is succeed", testCase{name: "pod2", namespace: "default", want: true}),
			Entry("Pod is failed", testCase{name: "pod3", namespace: "default", want: true}),
			Entry("Pod's deletion timestamp not nil", testCase{name: "pod4", namespace: "default", want: true}),
		)
	})

	Describe("GetPodByName", func() {
		var (
			namespace  string
			fakeClient client.Client
		)

		BeforeEach(func() {
			namespace = "default"
			pods := []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod3", Namespace: namespace},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: corev1.PodFailed},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod4",
						Namespace:         namespace,
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
						Finalizers:        []string{"test.finalizer"},
					},
					Spec: corev1.PodSpec{},
				},
			}

			testPods := []runtime.Object{}
			for _, pod := range pods {
				testPods = append(testPods, pod.DeepCopy())
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testPods...)
		})

		It("should return nil for non-existent pod", func() {
			pod, _ := GetPodByName(fakeClient, "notExist", namespace)
			Expect(pod).To(BeNil())
		})

		It("should return pod for existing pod", func() {
			pod, _ := GetPodByName(fakeClient, "pod1", namespace)
			Expect(pod).NotTo(BeNil())
			Expect(pod.Name).To(Equal("pod1"))
			Expect(pod.Namespace).To(Equal(namespace))
		})
	})

	Describe("IsSucceededPod", func() {
		DescribeTable("should correctly identify succeeded pods",
			func(phase corev1.PodPhase, want bool) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "testPod", Namespace: "default"},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: phase},
				}
				Expect(IsSucceededPod(pod)).To(Equal(want))
			},
			Entry("running pod", corev1.PodRunning, false),
			Entry("succeeded pod", corev1.PodSucceeded, true),
			Entry("failed pod", corev1.PodFailed, false),
		)
	})

	Describe("IsFailedPod", func() {
		DescribeTable("should correctly identify failed pods",
			func(phase corev1.PodPhase, want bool) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "testPod", Namespace: "default"},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: phase},
				}
				Expect(IsFailedPod(pod)).To(Equal(want))
			},
			Entry("running pod", corev1.PodRunning, false),
			Entry("succeeded pod", corev1.PodSucceeded, false),
			Entry("failed pod", corev1.PodFailed, true),
		)
	})

	Describe("isRunningAndReady", func() {
		DescribeTable("should correctly identify running and ready pods",
			func(phase corev1.PodPhase, conditions []corev1.PodCondition, want bool) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "testPod", Namespace: "default"},
					Spec:       corev1.PodSpec{},
					Status:     corev1.PodStatus{Phase: phase, Conditions: conditions},
				}
				Expect(isRunningAndReady(pod)).To(Equal(want))
			},
			Entry("running and ready pod", corev1.PodRunning, []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			}, true),
			Entry("succeeded pod", corev1.PodSucceeded, nil, false),
			Entry("failed pod", corev1.PodFailed, nil, false),
		)
	})

	Describe("MergeNodeSelectorAndNodeAffinity", func() {
		type testCase struct {
			nodeSelector map[string]string
			podAffinity  *corev1.Affinity
			want         *corev1.NodeAffinity
		}

		DescribeTable("should correctly merge node selector and node affinity",
			func(tc testCase) {
				nodeAffinity := MergeNodeSelectorAndNodeAffinity(tc.nodeSelector, tc.podAffinity)
				Expect(nodeAffinity).To(Equal(tc.want))
			},
			Entry("pod affinity nil", testCase{
				nodeSelector: nil,
				podAffinity:  nil,
				want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{},
				}},
			}),
			Entry("node affinity in pod nil", testCase{
				nodeSelector: nil,
				podAffinity:  &corev1.Affinity{},
				want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{},
				}},
			}),
			Entry("node affinity in pod is empty", testCase{
				nodeSelector: nil,
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
				},
				want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{},
				}},
			}),
			Entry("no exist node affinity with preferred scheduling", testCase{
				nodeSelector: map[string]string{"a": "b"},
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
							{
								Preference: corev1.NodeSelectorTerm{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{Key: "c", Operator: corev1.NodeSelectorOpIn, Values: []string{"d"}},
									},
								},
							},
						},
					},
				},
				want: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "a", Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "c", Operator: corev1.NodeSelectorOpIn, Values: []string{"d"}},
								},
							},
						},
					},
				},
			}),
			Entry("no exist node affinity simple", testCase{
				nodeSelector: map[string]string{"a": "b"},
				podAffinity:  nil,
				want: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "a", Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}},
								},
							},
						},
					},
				},
			}),
			Entry("exist node affinity", testCase{
				nodeSelector: map[string]string{"a": "b"},
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{Key: "c", Operator: corev1.NodeSelectorOpIn, Values: []string{"d"}},
									},
								},
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{Key: "e", Operator: corev1.NodeSelectorOpIn, Values: []string{"f"}},
									},
								},
							},
						},
					},
				},
				want: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "c", Operator: corev1.NodeSelectorOpIn, Values: []string{"d"}},
									{Key: "a", Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}},
								},
							},
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "e", Operator: corev1.NodeSelectorOpIn, Values: []string{"f"}},
									{Key: "a", Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}},
								},
							},
						},
					},
				},
			}),
		)
	})
})
