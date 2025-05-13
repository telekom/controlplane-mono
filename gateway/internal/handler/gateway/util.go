package gateway

import (
	"context"

	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/client"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	v1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func GetGatewayByRef(ctx context.Context, ref types.ObjectRef) (bool, *v1.Gateway, error) {
	client := client.ClientFromContextOrDie(ctx)

	gateway := &v1.Gateway{}
	err := client.Get(ctx, ref.K8s(), gateway)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, errors.Wrapf(err, "failed to get gateway %s", ref.String())
	}
	if !meta.IsStatusConditionTrue(gateway.GetConditions(), condition.ConditionTypeReady) {
		return false, nil, nil
	}
	return true, gateway, nil
}
