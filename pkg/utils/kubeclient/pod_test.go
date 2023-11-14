package kubeclient

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

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

func TestGetPVCNamesFromPod(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pod := corev1.Pod{}
	var pvcNamesWant []string
	for i := 1; i <= 30; i++ {
		switch rand.Intn(4) {
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

	if !reflect.DeepEqual(pvcNames, pvcNamesWant) {
		t.Errorf("the result of GetPVCNamesFromPod is not right")
	}

}

func TestIsCompletePod(t *testing.T) {
	namespace := "default"
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod1",
		Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod2",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod3",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod4",
			Namespace:         namespace,
			DeletionTimestamp: &metav1.Time{Time: time.Now()}},
		Spec: corev1.PodSpec{},
	}}

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Pod doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want: false,
		},
		{
			name: "Pod is running",
			args: args{
				name:      "pod1",
				namespace: namespace,
			},
			want: false,
		},
		{
			name: "Pod is succeed",
			args: args{
				name:      "pod2",
				namespace: namespace,
			},
			want: true,
		}, {
			name: "Pod is failed",
			args: args{
				name:      "pod3",
				namespace: namespace,
			},
			want: true,
		}, {
			name: "Pod's deletion timestamp not nil",
			args: args{
				name:      "pod4",
				namespace: namespace,
			},
			want: true,
		},
	}

	testPods := []runtime.Object{}

	for _, pod := range pods {
		testPods = append(testPods, pod.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPods...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pod corev1.Pod
			var podToTest *corev1.Pod
			key := types.NamespacedName{
				Namespace: tt.args.namespace,
				Name:      tt.args.name,
			}
			_ = client.Get(context.TODO(), key, &pod)
			// if err != nil {
			// 	t.Errorf("testcase %v IsCompletePod() got err: %v", tt.name, err.Error())
			// }
			if len(pod.Name) == 0 {
				podToTest = nil
			} else {
				podToTest = &pod
			}

			if got := IsCompletePod(podToTest); got != tt.want {
				t.Errorf("testcase %v IsCompletePod() = %v, want %v, pod %v", tt.name, got, tt.want, pod)
			}
		})
	}
}

func TestGetPodByName(t *testing.T) {
	namespace := "default"
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod1",
		Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod2",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod3",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod4",
			Namespace:         namespace,
			DeletionTimestamp: &metav1.Time{Time: time.Now()}},
		Spec: corev1.PodSpec{},
	}}

	testPods := []runtime.Object{}

	for _, pod := range pods {
		testPods = append(testPods, pod.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPods...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Pod
	}{
		{
			name: "Pod doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want: nil,
		},
		{
			name: "Pod is running",
			args: args{
				name:      "pod1",
				namespace: namespace,
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1",
					Namespace: namespace},
				Spec: corev1.PodSpec{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod, _ := GetPodByName(client, tt.args.name, tt.args.namespace)
			if tt.want == nil {
				if pod != nil {
					t.Errorf("testcase %v GetPodByName() = %v, want %v", tt.name, pod, tt.want)
				}
			} else {
				if pod == nil {
					t.Errorf("testcase %v GetPodByName() = %v, want %v", tt.name, pod, tt.want)
				} else if pod.Name != tt.args.name || pod.Namespace != tt.args.namespace {
					t.Errorf("testcase %v GetPodByName() = %v, want %v", tt.name, pod, tt.want)
				}
			}

		})
	}
}

func TestIsSucceededPod(t *testing.T) {
	namespace := "default"
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "runningPod",
		Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "succeedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "failedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}}
	type args struct {
		pod *corev1.Pod
	}
	type testcase struct {
		name string
		args args
		want bool
	}

	tests := []testcase{}

	for _, pod := range pods {
		tests = append(tests, testcase{
			name: pod.Name,
			args: args{
				pod: pod,
			},
		})
	}

	tests[0].want = false
	tests[1].want = true
	tests[2].want = false

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSucceededPod(tt.args.pod); got != tt.want {
				t.Errorf("testcase %v IsSucceededPod() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsFailedPod(t *testing.T) {
	namespace := "default"
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "runningPod",
		Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "succeedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "failedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}}
	type args struct {
		pod *corev1.Pod
	}
	type testcase struct {
		name string
		args args
		want bool
	}

	tests := []testcase{}

	for _, pod := range pods {
		tests = append(tests, testcase{
			name: pod.Name,
			args: args{
				pod: pod,
			},
		})
	}

	tests[0].want = false
	tests[1].want = false
	tests[2].want = true

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFailedPod(tt.args.pod); got != tt.want {
				t.Errorf("testcase %v IsFailedPod() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsRunningAndReady(t *testing.T) {

	namespace := "default"
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "runningPod",
		Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "succeedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "failedPod",
			Namespace: namespace},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}}
	type args struct {
		pod *corev1.Pod
	}
	type testcase struct {
		name      string
		args      args
		isRunning bool
	}

	tests := []testcase{}

	for _, pod := range pods {
		tests = append(tests, testcase{
			name: pod.Name,
			args: args{
				pod: pod,
			},
		})
	}

	tests[0].isRunning = true
	tests[1].isRunning = false
	tests[2].isRunning = false

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRunningAndReady(tt.args.pod); got != tt.isRunning {
				t.Errorf("testcase %v isRunningAndReady() = %v, want %v", tt.name, got, tt.isRunning)
			}
		})
	}
}

func TestMergeNodeSelectorAndNodeAffinity(t *testing.T) {
	type args struct {
		nodeSelector map[string]string
		podAffinity  *corev1.Affinity
	}
	tests := []struct {
		name string
		args args
		want *corev1.NodeAffinity
	}{
		{
			name: "pod affinity nil",
			args: args{},
			want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{},
			}},
		},
		{
			name: "node affinity in pod nil",
			args: args{
				podAffinity: &corev1.Affinity{},
			},
			want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{},
			}},
		},
		{
			name: "node affinity in pod is empty",
			args: args{
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
				},
			},
			want: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{},
			}},
		},
		{
			name: "no exist node affinity",
			args: args{
				nodeSelector: map[string]string{
					"a": "b",
				},
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
							{
								Preference: corev1.NodeSelectorTerm{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "c",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"d"},
										},
									},
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
								{
									Key:      "a",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"b"},
								},
							},
						},
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
					{
						Preference: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "c",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"d"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "no exist node affinity",
			args: args{
				nodeSelector: map[string]string{
					"a": "b",
				},
			},
			want: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "a",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"b"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "exist node affinity",
			args: args{
				nodeSelector: map[string]string{
					"a": "b",
				},
				podAffinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "c",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"d"},
										},
									},
								},
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "e",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"f"},
										},
									},
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
								{
									Key:      "c",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"d"},
								},
								{
									Key:      "a",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"b"},
								},
							},
						},
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "e",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"f"},
								},
								{
									Key:      "a",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"b"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeAffinity := MergeNodeSelectorAndNodeAffinity(tt.args.nodeSelector, tt.args.podAffinity)
			if !reflect.DeepEqual(nodeAffinity, tt.want) {
				t.Errorf("testcase %v IsFailedPod() = %v, want %v", tt.name, nodeAffinity, tt.want)
			}
		})
	}
}
