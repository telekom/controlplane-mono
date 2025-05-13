package consumer

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/handler"
	v1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"github.com/telekom/controlplane-mono/gateway/internal/handler/gateway"
	"github.com/telekom/controlplane-mono/gateway/internal/handler/realm"
	"github.com/telekom/controlplane-mono/gateway/pkg/kongutil"
	secrets "github.com/telekom/controlplane-mono/secret-manager/pkg/api"
)

var _ handler.Handler[*v1.Consumer] = &ConsumerHandler{}

type ConsumerHandler struct{}

func (h *ConsumerHandler) CreateOrUpdate(ctx context.Context, consumer *v1.Consumer) error {

	consumer.SetCondition(condition.NewProcessingCondition("Processing", "Processing consumer"))
	consumer.SetCondition(condition.NewNotReadyCondition("ConsumerNotReady", "Consumer not ready"))

	ready, realm, err := realm.GetRealmByRef(ctx, consumer.Spec.Realm)
	if err != nil {
		return err
	}
	if !ready {
		consumer.SetCondition(condition.NewBlockedCondition("Realm not ready"))
		consumer.SetCondition(condition.NewNotReadyCondition("RealmNotReady", "Realm not ready"))
		return nil
	}

	ready, gateway, err := gateway.GetGatewayByRef(ctx, *realm.Spec.Gateway)
	if err != nil {
		return err
	}
	if !ready {
		consumer.SetCondition(condition.NewBlockedCondition("Gateway not ready"))
		consumer.SetCondition(condition.NewNotReadyCondition("GatewayNotReady", "Gateway not ready"))
		return nil
	}

	gateway.Spec.Admin.ClientSecret, err = secrets.Get(ctx, gateway.Spec.Admin.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "failed to get gateway client secret")
	}

	kc, err := kongutil.GetClientFor(gateway)
	if err != nil {
		return errors.Wrap(err, "failed to get kong client")
	}

	err = kc.CreateOrReplaceConsumer(ctx, consumer.Spec.Name)
	if err != nil {
		return errors.Wrap(err, "failed to create or update consumer")
	}

	consumer.SetCondition(condition.NewDoneProcessingCondition("Consumer is ready"))
	consumer.SetCondition(condition.NewReadyCondition("ConsumerReady", "Consumer is ready"))

	return nil
}

func (h *ConsumerHandler) Delete(ctx context.Context, consumer *v1.Consumer) error {
	log := logr.FromContextOrDiscard(ctx)
	found, realm, err := realm.GetRealmByRef(ctx, consumer.Spec.Realm)
	if err != nil {
		return err
	}
	if !found {
		log.Info("Realm not found, skipping consumer deletion")
		return nil
	}

	found, gateway, err := gateway.GetGatewayByRef(ctx, *realm.Spec.Gateway)
	if err != nil {
		return err
	}
	if !found {
		log.Info("Gateway not found, skipping consumer deletion")
		return nil
	}

	kc, err := kongutil.GetClientFor(gateway)
	if err != nil {
		return errors.Wrap(err, "failed to get kong client")
	}

	err = kc.DeleteConsumer(ctx, consumer.Spec.Name)
	if err != nil {
		return errors.Wrap(err, "failed to create or update consumer")
	}

	return nil
}
