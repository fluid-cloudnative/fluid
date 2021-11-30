package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetAlluxioRuntime(t *testing.T) {
	runtimeNamespace := "default"
	runtimeName := "alluxio-runtime-1"
	alluxio := &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: runtimeNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, alluxio)

	fakeClient := fake.NewFakeClientWithScheme(s, alluxio)

	tests := []struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		{
			name:      runtimeName,
			namespace: runtimeNamespace,
			wantName:  runtimeName,
			notFound:  false,
		},
		{
			name:      runtimeName + "not-exist",
			namespace: runtimeNamespace,
			wantName:  "",
			notFound:  true,
		},
		{
			name:      runtimeName,
			namespace: runtimeNamespace + "not-exist",
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range tests {
		gotRuntime, err := GetAlluxioRuntime(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || gotRuntime != nil {
				t.Errorf("%d check failure, want to got nil", k)
			} else {
				if !apierrs.IsNotFound(err) {
					t.Errorf("%d check failure, want notFound err but got %s", k, err)
				}
			}
		} else {
			if gotRuntime.Name != item.wantName {
				t.Errorf("%d check failure, got AlluxioRuntime name: %s, want name: %s", k, gotRuntime.Name, item.wantName)
			}
		}
	}
}

func TestGetJuiceFSRuntime(t *testing.T) {
	runtimeNamespace := "default"
	runtimeName := "juicefs-runtime-1"
	juicefsRuntime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: runtimeNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, juicefsRuntime)

	fakeClient := fake.NewFakeClientWithScheme(s, juicefsRuntime)

	tests := []struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		{
			name:      runtimeName,
			namespace: runtimeNamespace,
			wantName:  runtimeName,
			notFound:  false,
		},
		{
			name:      runtimeName + "not-exist",
			namespace: runtimeNamespace,
			wantName:  "",
			notFound:  true,
		},
		{
			name:      runtimeName,
			namespace: runtimeNamespace + "not-exist",
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range tests {
		gotRuntime, err := GetJuiceFSRuntime(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || gotRuntime != nil {
				t.Errorf("%d check failure, want to got nil", k)
			} else {
				if !apierrs.IsNotFound(err) {
					t.Errorf("%d check failure, want notFound err but got %s", k, err)
				}
			}
		} else {
			if gotRuntime.Name != item.wantName {
				t.Errorf("%d check failure, got JuiceFSRuntime name: %s, want name: %s", k, gotRuntime.Name, item.wantName)
			}
		}
	}
}

func TestGetJindoRuntime(t *testing.T) {
	runtimeNamespace := "default"
	runtimeName := "jindo-runtime-1"
	jindo := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: runtimeNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, jindo)

	fakeClient := fake.NewFakeClientWithScheme(s, jindo)

	tests := []struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		{
			name:      runtimeName,
			namespace: runtimeNamespace,
			wantName:  runtimeName,
			notFound:  false,
		},
		{
			name:      runtimeName + "not-exist",
			namespace: runtimeNamespace,
			wantName:  "",
			notFound:  true,
		},
		{
			name:      runtimeName,
			namespace: runtimeNamespace + "not-exist",
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range tests {
		gotRuntime, err := GetJindoRuntime(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || gotRuntime != nil {
				t.Errorf("%d check failure, want to got nil", k)
			} else {
				if !apierrs.IsNotFound(err) {
					t.Errorf("%d check failure, want notFound err but got %s", k, err)
				}
			}
		} else {
			if gotRuntime.Name != item.wantName {
				t.Errorf("%d check failure, got AlluxioRuntime name: %s, want name: %s", k, gotRuntime.Name, item.wantName)
			}
		}
	}
}

func TestGetGooseFSRuntime(t *testing.T) {
	runtimeNamespace := "default"
	runtimeName := "goosefs-runtime-1"
	goosefs := &datav1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: runtimeNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, goosefs)

	fakeClient := fake.NewFakeClientWithScheme(s, goosefs)

	tests := []struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		{
			name:      runtimeName,
			namespace: runtimeNamespace,
			wantName:  runtimeName,
			notFound:  false,
		},
		{
			name:      runtimeName + "not-exist",
			namespace: runtimeNamespace,
			wantName:  "",
			notFound:  true,
		},
		{
			name:      runtimeName,
			namespace: runtimeNamespace + "not-exist",
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range tests {
		gotRuntime, err := GetGooseFSRuntime(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || gotRuntime != nil {
				t.Errorf("%d check failure, want to got nil", k)
			} else {
				if !apierrs.IsNotFound(err) {
					t.Errorf("%d check failure, want notFound err but got %s", k, err)
				}
			}
		} else {
			if gotRuntime.Name != item.wantName {
				t.Errorf("%d check failure, got GooseFSRuntime name: %s, want name: %s", k, gotRuntime.Name, item.wantName)
			}
		}
	}
}

func TestAddRuntimesIfNotExist(t *testing.T) {
	var runtime1 = datav1alpha1.Runtime{
		Name:     "imagenet",
		Category: common.AccelerateCategory,
	}
	var runtime2 = datav1alpha1.Runtime{
		Name:     "mock-name",
		Category: "mock-category",
	}
	var runtime3 = datav1alpha1.Runtime{
		Name:     "cifar10",
		Category: common.AccelerateCategory,
	}
	var testCases = []struct {
		description string
		runtimes    []datav1alpha1.Runtime
		newRuntime  datav1alpha1.Runtime
		expected    []datav1alpha1.Runtime
	}{
		{"add runtime to an empty slices successfully",
			[]datav1alpha1.Runtime{}, runtime1, []datav1alpha1.Runtime{runtime1}},
		{"duplicate runtime will not be added",
			[]datav1alpha1.Runtime{runtime1}, runtime1, []datav1alpha1.Runtime{runtime1}},
		{"add runtime of different name and category successfully",
			[]datav1alpha1.Runtime{runtime1}, runtime2, []datav1alpha1.Runtime{runtime1, runtime2}},
		{"runtime of the same category but different name will not be added",
			[]datav1alpha1.Runtime{runtime1}, runtime3, []datav1alpha1.Runtime{runtime1}},
	}
	var runtimeSliceEqual = func(a, b []datav1alpha1.Runtime) bool {
		if len(a) != len(b) || (a == nil) != (b == nil) {
			return false
		}
		for i, s := range a {
			if s != b[i] {
				return false
			}
		}
		return true
	}
	for _, tc := range testCases {
		if updatedRuntimes := AddRuntimesIfNotExist(tc.runtimes, tc.newRuntime); !runtimeSliceEqual(tc.expected, updatedRuntimes) {
			t.Errorf("%s, expected %#v, got %#v",
				tc.description, tc.expected, updatedRuntimes)
		}
	}
}
