package utils

import (
	"reflect"
	"testing"
)

func TestFillSliceWithString(t *testing.T) {
	type args struct {
		str string
		num int
	}
	tests := []struct {
		name string
		args args
		want *[]string
	}{
		{
			name: "Fill Slice Test1",
			args: args{
				str: "foo",
				num: 3,
			},
			want: &[]string{"foo", "foo", "foo"},
		},
		{
			name: "Fill Slice Test2",
			args: args{
				str: "bar",
				num: 0,
			},
			want: &[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FillSliceWithString(tt.args.str, tt.args.num); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FillSliceWithString() = %v, want %v", got, tt.want)
			}
		})
	}
}
