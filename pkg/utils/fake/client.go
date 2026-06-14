/*
Copyright 2022 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ContextAwareClient wraps a fake client and returns ctx.Err() before delegating.
type ContextAwareClient struct {
	client.Client
}

func (c ContextAwareClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c ContextAwareClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.List(ctx, list, opts...)
}

func (c ContextAwareClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c ContextAwareClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Delete(ctx, obj, opts...)
}

func (c ContextAwareClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Update(ctx, obj, opts...)
}

func (c ContextAwareClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c ContextAwareClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.DeleteAllOf(ctx, obj, opts...)
}

// NewFakeClientWithScheme is to fix the issue by wrappering it:
// fake.NewFakeClientWithScheme is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClientWithScheme(clientScheme *runtime.Scheme, initObjs ...runtime.Object) client.Client {
	var clientObjs []client.Object
	for _, obj := range initObjs {
		clientObj, ok := obj.(client.Object)
		if ok {
			clientObjs = append(clientObjs, clientObj)
		}
	}

	return fake.NewClientBuilder().WithScheme(clientScheme).WithRuntimeObjects(initObjs...).WithStatusSubresource(clientObjs...).Build()
}

// NewFakeClient is to fix the issue by wrappering it:
// fake.NewFakeClient is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClient(initObjs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
