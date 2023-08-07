package patch

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPatchBodyGeneration(t *testing.T) {
	patchReq := NewStrategicPatch().
		AddFinalizer("new-finalizer").
		RemoveFinalizer("old-finalizer").
		InsertLabel("new-label", "foo1").
		DeleteLabel("old-label").
		InsertAnnotation("new-annotation", "foo2").
		DeleteAnnotation("old-annotation")

	expectedPatchBody := fmt.Sprintf(`{"metadata":{"labels":{"new-label":"foo1","old-label":null},"annotations":{"new-annotation":"foo2","old-annotation":null},"finalizers":["new-finalizer"],"$deleteFromPrimitiveList/finalizers":["old-finalizer"]}}`)

	if !reflect.DeepEqual(patchReq.String(), expectedPatchBody) {
		t.Fatalf("Not equal: \n%s \n%s", expectedPatchBody, patchReq.String())
	}

}

func TestMergePatchBody(t *testing.T) {
	finalizers := []string{"origin-finalizer"}
	tests := []struct {
		name                   string
		mergePatchReq          *Patch
		expectedMergePatchBody string
	}{
		{
			name: "add finalizer",
			mergePatchReq: NewMergePatch().OverrideFinalizer(append(finalizers, "new-finalizer")).
				InsertLabel("new-label", "foo1").DeleteLabel("old-label").
				InsertAnnotation("new-annotation", "foo2").DeleteAnnotation("old-annotation"),
			expectedMergePatchBody: fmt.Sprintf(`{"metadata":{"labels":{"new-label":"foo1","old-label":null},"annotations":{"new-annotation":"foo2","old-annotation":null},"finalizers":["origin-finalizer","new-finalizer"]}}`),
		},
		{
			name: "remove finalizer",
			mergePatchReq: NewMergePatch().OverrideFinalizer(append(finalizers, "new-finalizer")).
				InsertLabel("new-label", "foo1").DeleteLabel("old-label").
				InsertAnnotation("new-annotation", "foo2").DeleteAnnotation("old-annotation"),
			expectedMergePatchBody: fmt.Sprintf(`{"metadata":{"labels":{"new-label":"foo1","old-label":null},"annotations":{"new-annotation":"foo2","old-annotation":null},"finalizers":["origin-finalizer","new-finalizer"]}}`),
		},
		{
			name: "add finalizer",
			mergePatchReq: NewMergePatch().OverrideFinalizer(make([]string, 0)).
				InsertLabel("new-label", "foo1").DeleteLabel("old-label").
				InsertAnnotation("new-annotation", "foo2").DeleteAnnotation("old-annotation"),
			expectedMergePatchBody: fmt.Sprintf(`{"metadata":{"labels":{"new-label":"foo1","old-label":null},"annotations":{"new-annotation":"foo2","old-annotation":null},"finalizers":null}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.mergePatchReq.String(), tt.expectedMergePatchBody) {
				t.Fatalf("Not equal: \n%s \n%s", tt.expectedMergePatchBody, tt.mergePatchReq.String())
			}
		})
	}

}

func TestPodPatchOperations(t *testing.T) {
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "main",
					Image: "busybox",
				},
			},
		},
	}
	nsName := types.NamespacedName{Name: "hello", Namespace: "default"}

	fake := fakeclient.NewFakeClientWithScheme(scheme.Scheme)
	_ = fake.Create(context.Background(), p)

	patch := NewStrategicPatch()
	patch.InsertLabel("foo", "bar")
	err := fake.Patch(context.Background(), p, patch)
	assert.Nil(t, err)
	err = fake.Get(context.Background(), nsName, p)
	assert.Nil(t, err)
	assert.Equal(t, p.Labels["foo"], "bar")
	patch.InsertLabel("foo", "bar1")
	err = fake.Patch(context.Background(), p, patch)
	assert.Nil(t, err)
	err = fake.Get(context.Background(), nsName, p)
	assert.Nil(t, err)
	assert.Equal(t, p.Labels["foo"], "bar1")

	patch = NewStrategicPatch()
	patch.InsertAnnotation("foo", "bar")
	err = fake.Patch(context.Background(), p, patch)
	assert.Nil(t, err)
	err = fake.Get(context.Background(), nsName, p)
	assert.Nil(t, err)
	assert.Equal(t, p.Annotations["foo"], "bar")

	patch = NewStrategicPatch()
	patch.AddFinalizer("finalizer")
	err = fake.Patch(context.Background(), p, patch)
	assert.Nil(t, err)
	err = fake.Get(context.Background(), nsName, p)
	assert.Nil(t, err)
	assert.Equal(t, len(p.Finalizers), 1)
	assert.Equal(t, p.Finalizers[0], "finalizer")
}
