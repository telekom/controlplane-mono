package controller

import "github.com/telekom/controlplane-mono/secret-manager/pkg/backend"

type Controller interface {
	SecretsController
	OnboardController
}

type controller struct {
	SecretsController
	OnboardController
}

func NewController[T backend.SecretId, S backend.Secret[T]](b backend.Backend[T, S], o backend.Onboarder) Controller {
	return &controller{
		SecretsController: NewSecretsController(b),
		OnboardController: NewOnboardController(o),
	}
}
