package utils

import (
	"reflect"
	"testing"
)

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

func TestUnionMapsWithOverride(t *testing.T) {
	type args struct {
		map1 map[string]string
		map2 map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "no_shared_elements",
			args: args{
				map1: map[string]string{"key1": "value1", "key2": "value2"},
				map2: map[string]string{"keyA": "valueA", "keyB": "valueB"},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"keyA": "valueA",
				"keyB": "valueB",
			},
		},
		{
			name: "with_shared_elements",
			args: args{
				map1: map[string]string{"key1": "value1", "key2": "value2"},
				map2: map[string]string{"key2": "new_value", "key3": "value3"},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "new_value",
				"key3": "value3",
			},
		},
		{
			name: "nil_map",
			args: args{
				map1: map[string]string{"key1": "value1", "key2": "value2"},
				map2: nil,
			},
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "nil_maps",
			args: args{
				map1: nil,
				map2: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnionMapsWithOverride(tt.args.map1, tt.args.map2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnionMapsWithOverride() = %v, want %v", got, tt.want)
			}
		})
	}
}
