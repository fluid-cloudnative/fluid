package kubeclient

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestGetPVCNamesFromPod(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pod := v1.Pod{}
	var pvcNamesWant []string
	for i := 1; i <= 30; i++ {
		switch rand.Intn(4) {
		case 0:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/tmp/data" + strconv.Itoa(i),
					},
				},
			})
		case 1:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: "pvc" + strconv.Itoa(i),
						ReadOnly:  true,
					},
				},
			})
			pvcNamesWant = append(pvcNamesWant, "pvc"+strconv.Itoa(i))
		case 2:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			})
		case 3:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					NFS: &v1.NFSVolumeSource{
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
	pods := []*v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod1",
		Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod2",
			Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod3",
			Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod4",
			Namespace:         namespace,
			DeletionTimestamp: &metav1.Time{Time: time.Now()}},
		Spec: v1.PodSpec{},
	}}

	type args struct {
		name        string
		namespace   string
		annotations map[string]string
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
			var pod v1.Pod
			key := types.NamespacedName{
				Namespace: tt.args.namespace,
				Name:      tt.args.name,
			}
			client.Get(context.TODO(), key, &pod)
			// if err != nil {
			// 	t.Errorf("testcase %v IsCompletePod() got err: %v", tt.name, err.Error())
			// }
			if got := IsCompletePod(&pod); got != tt.want {
				t.Errorf("testcase %v IsCompletePod() = %v, want %v, pod %v", tt.name, got, tt.want, pod)
			}
		})
	}
}

func TestGetPodByName(t *testing.T) {
	namespace := "default"
	pods := []*v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod1",
		Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod2",
			Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod3",
			Namespace: namespace},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "pod4",
			Namespace:         namespace,
			DeletionTimestamp: &metav1.Time{Time: time.Now()}},
		Spec: v1.PodSpec{},
	}}

	testPods := []runtime.Object{}

	for _, pod := range pods {
		testPods = append(testPods, pod.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPods...)

	type args struct {
		name        string
		namespace   string
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want *v1.Pod
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
			want: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1",
					Namespace: namespace},
				Spec: v1.PodSpec{},
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
