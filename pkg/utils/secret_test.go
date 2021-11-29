package utils

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSecret(t *testing.T) {

	secretName := "mysecret"
	secretNamespace := "default"

	mockSecret1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, mockSecret1)

	testCases := map[string]struct {
		name          string
		namespace     string
		wantName      string
		wantNamespace string
		notFound      bool
	}{
		"Case 1: get existed secret": {
			name:          secretName,
			namespace:     secretNamespace,
			wantName:      secretName,
			wantNamespace: secretNamespace,
			notFound:      false,
		},
		"Case 2: get non-existed secret": {
			name:          secretName + "not-exist",
			namespace:     secretNamespace,
			wantName:      "",
			wantNamespace: "",
			notFound:      true,
		},
	}

	for caseName, item := range testCases {
		gotSecret, err := GetSecret(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || IgnoreNotFound(err) != nil {
				t.Errorf("%s check failure, want not found error, but got %v", caseName, err)
			}
		} else {
			if gotSecret == nil {
				t.Errorf("%s check failure, got nil secret", caseName)
			} else if gotSecret.Name != item.wantName || gotSecret.Namespace != item.wantNamespace {
				t.Errorf("%s check failure, want secret with name %s and namespace %s, but got name %s and namespace %s",
					caseName,
					gotSecret.Name,
					gotSecret.Namespace,
					item.wantName,
					item.wantNamespace)
			}
		}
	}
}
