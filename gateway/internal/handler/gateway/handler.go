package gateway

import (
	"context"

	"github.com/pkg/errors"
	cc "github.com/telekom/controlplane-mono/common/pkg/client"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/handler"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ handler.Handler[*gatewayv1.Gateway] = &GatewayHandler{}

type GatewayHandler struct{}

func (h *GatewayHandler) CreateOrUpdate(ctx context.Context, gw *gatewayv1.Gateway) error {

	gw.SetCondition(condition.NewDoneProcessingCondition("Created Gateway"))
	gw.SetCondition(condition.NewReadyCondition("Ready", "Gateway is ready"))
	return nil
}

func (h *GatewayHandler) Delete(ctx context.Context, object *gatewayv1.Gateway) error {
	c := cc.ClientFromContextOrDie(ctx)

	// If the Gateway which is referenced by the realms is deleted, we need to delete
	// the realms as well.
	realms := &gatewayv1.RealmList{}
	err := c.List(ctx, realms, client.InNamespace(object.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to list realms")
	}

	for _, realm := range realms.Items {
		if realm.Spec.Gateway.Equals(object) {
			err := c.Delete(ctx, &realm)
			if err != nil {
				return errors.Wrapf(err, "failed to delete realm %s", realm.Name)
			}
		}
	}

	return nil
}
