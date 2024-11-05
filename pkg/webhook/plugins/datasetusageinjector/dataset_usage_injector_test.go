package datasetusageinjector

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDatasetUsageInjector_Mutate(t *testing.T) {
	runtimeInfo1, err := base.BuildRuntimeInfo("demo-dataset-1", "fluid-test", "")
	if err != nil {
		t.Fatalf("build runtime info failed with %v", err)
	}
	runtimeInfo2, err := base.BuildRuntimeInfo("demo-dataset-2", "fluid-test", "")
	if err != nil {
		t.Fatalf("build runtime info failed with %v", err)
	}

	type args struct {
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
	}
	tests := []struct {
		name    string
		args    args
		wantPod *corev1.Pod
		wantErr bool
	}{
		{
			name: "one_dataset_mounted",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid-test",
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					"demo-dataset-1": runtimeInfo1,
				},
			},
			wantPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid-test",
					Annotations: map[string]string{
						common.LabelAnnotationDatasetsInUse: "demo-dataset-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple_datasets_mounted",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "fluid-test",
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					"demo-dataset-2": runtimeInfo2,
					"demo-dataset-1": runtimeInfo1,
				},
			},
			wantPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "fluid-test",
					Annotations: map[string]string{
						common.LabelAnnotationDatasetsInUse: "demo-dataset-1,demo-dataset-2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no_datasets_mounted",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-no-dataset",
						Namespace: "fluid-test",
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{},
			},
			wantPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-no-dataset",
					Namespace: "fluid-test",
				},
			},
			wantErr: false,
		},
	}

	plugin, err := NewPlugin(fake.NewFakeClient(), "")
	if err != nil {
		t.Fatalf("expected no error calling NewPlugin(), got %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := plugin.Mutate(tt.args.pod, tt.args.runtimeInfos)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MountedDatasetInjector.Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.pod, tt.wantPod) {
				t.Fatalf("DeepEqual not expected, diff between wantPod and expectedPod: %v", cmp.Diff(tt.wantPod, tt.args.pod))
			}
		})
	}
}
