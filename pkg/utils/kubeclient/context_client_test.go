package kubeclient

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type contextAwareClient struct {
	client.Client
}

func (c contextAwareClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c contextAwareClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c contextAwareClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Delete(ctx, obj, opts...)
}

func (c contextAwareClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Client.Update(ctx, obj, opts...)
}
