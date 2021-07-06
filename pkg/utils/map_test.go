package utils

import "testing"

func TestContainsAll(t *testing.T) {
	type args struct {
		m     map[string]string
		slice []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil slice",
			args: args{
				m: map[string]string{
					"foo": "1",
				},
				slice: nil,
			},
			want: true,
		},
		{
			name: "nil map",
			args: args{
				m:     nil,
				slice: []string{"foo", "bar"},
			},
			want: false,
		},
		{
			name: "contains all",
			args: args{
				m: map[string]string{
					"foo": "1",
					"bar": "2",
					"xxx": "3",
				},
				slice: []string{"foo", "xxx"},
			},
			want: true,
		},
		{
			name: "contains some",
			args: args{
				m: map[string]string{
					"foo": "1",
					"bar": "2",
					"xxx": "3",
				},
				slice: []string{"foo", "yyy"},
			},
			want: false,
		},
		{
			name: "contains none",
			args: args{
				m: map[string]string{
					"foo": "1",
					"bar": "2",
					"xxx": "3",
				},
				slice: []string{"aaa", "bbb"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAll(tt.args.m, tt.args.slice); got != tt.want {
				t.Errorf("ContainsAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
