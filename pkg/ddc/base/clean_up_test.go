package base

import (
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetRemaining(t *testing.T) {
	ttl := int32(10)
	testcase := []struct {
		name              string
		dataload          datav1alpha1.DataLoad
		dataoperationType datav1alpha1.OperationType
		validRemaining    bool
		wantErr           bool
	}{
		{
			name: "get remaining time successfully",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							LastProbeTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    true,
			wantErr:           false,
		},
		{
			name: "not set ttl",
			dataload: datav1alpha1.DataLoad{
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							LastProbeTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
		{
			name: "data operation not completion",
			dataload: datav1alpha1.DataLoad{
				Status: datav1alpha1.OperationStatus{},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
		{
			name: "get remaining time < 0",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
				Status: datav1alpha1.OperationStatus{
					Conditions: []datav1alpha1.Condition{
						{
							LastProbeTime: v1.NewTime(time.Now().Add(-20 * time.Second)),
						},
					},
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			validRemaining:    false,
			wantErr:           false,
		},
	}
	for _, test := range testcase {
		remaining, err := GetRemainingTimeToCleanUp(&test.dataload, &test.dataload.Status, test.dataoperationType)
		if test.validRemaining != (remaining != nil && *remaining > 0) {
			t.Errorf("GetRemaining want validRemaining %v, get remaining %v", test.validRemaining, remaining)
		}
		if test.wantErr != (err != nil) {
			t.Errorf("GetRemaining want error %v, get error %v", test.wantErr, err)
		}
	}
}

func TestGetTTL(t *testing.T) {
	ttl := int32(10)
	testcase := []struct {
		name              string
		dataload          datav1alpha1.DataLoad
		dataoperationType datav1alpha1.OperationType
		ttl               *int32
		wantErr           bool
	}{
		{
			name: "get ttl",
			dataload: datav1alpha1.DataLoad{
				Spec: datav1alpha1.DataLoadSpec{
					TTLSecondsAfterFinished: &ttl,
				},
			},
			dataoperationType: datav1alpha1.DataLoadType,
			ttl:               &ttl,
			wantErr:           false,
		},
		{
			name:              "no ttl",
			dataload:          datav1alpha1.DataLoad{},
			dataoperationType: datav1alpha1.DataLoadType,
			ttl:               nil,
			wantErr:           false,
		},
		{
			name:              "wrong data operation type",
			dataload:          datav1alpha1.DataLoad{},
			dataoperationType: datav1alpha1.DataMigrateType,
			ttl:               nil,
			wantErr:           true,
		},
	}
	for _, test := range testcase {
		ttl, err := GetTTL(&test.dataload, test.dataoperationType)
		if ttl != test.ttl {
			t.Errorf("Get wrong ttl value, want %v, get %v", test.ttl, ttl)
		}
		if test.wantErr != (err != nil) {
			t.Errorf("GetTTL want error %v, get error %v", test.wantErr, err)
		}
	}
}
