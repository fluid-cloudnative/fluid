package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewFakeClientWithScheme is to fix the issue by wrappering it:
// fake.NewFakeClientWithScheme is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClientWithScheme(clientScheme *runtime.Scheme, initObjs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(clientScheme).WithRuntimeObjects(initObjs...).Build()
}

// NewFakeClient is to fix the issue by wrappering it:
// fake.NewFakeClient is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClient(initObjs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
