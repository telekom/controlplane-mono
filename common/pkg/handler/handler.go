package handler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler[T client.Object] interface {
	CreateOrUpdate(ctx context.Context, obj T) error
	Delete(ctx context.Context, obj T) error
}
