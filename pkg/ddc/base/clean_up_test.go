package base

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

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
