package kubeclient

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetCronJobStatus(t *testing.T) {
	nowTime := time.Now()
	testDate := metav1.NewTime(time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), 0, 0, 0, nowTime.Location()))

	namespace := "default"
	testCronJobInputs := []*batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: namespace,
			},
			Status: batchv1.CronJobStatus{
				LastScheduleTime: &testDate,
			},
		},
	}

	testCronJobs := []runtime.Object{}

	for _, cj := range testCronJobInputs {
		testCronJobs = append(testCronJobs, cj.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testCronJobs...)

	type args struct {
		key types.NamespacedName
	}
	tests := []struct {
		name    string
		args    args
		want    *batchv1.CronJobStatus
		wantErr bool
	}{
		{
			name: "CronJob exists",
			args: args{
				key: types.NamespacedName{
					Namespace: namespace,
					Name:      "test1",
				},
			},
			want: &batchv1.CronJobStatus{
				LastScheduleTime: &testDate,
			},
			wantErr: false,
		},
		{
			name: "CronJob exists",
			args: args{
				key: types.NamespacedName{
					Namespace: namespace,
					Name:      "test-notexist",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	patch := gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
		return true
	})
	defer patch.Reset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCronJobStatus(client, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCronJobStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCronJobStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
