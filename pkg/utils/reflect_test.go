package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestMatchFieldByType(t *testing.T) {

	pod := corev1.Pod{}

	targetType := reflect.TypeOf(corev1.Pod{})

	match := matchFieldByType(reflect.ValueOf(pod), targetType)

	if !match {
		t.Errorf("expect %v's type match %v", pod, targetType)
	}
}

func TestFieldNameByType(t *testing.T) {

}
