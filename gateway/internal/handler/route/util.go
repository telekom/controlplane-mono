package route

import (
	"context"

	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/client"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	v1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"k8s.io/apimachinery/pkg/api/meta"
)

func GetRouteByRef(ctx context.Context, ref types.ObjectRef) (*v1.Route, error) {
	client, _ := client.ClientFromContext(ctx)

	route := &v1.Route{}
	err := client.Get(context.Background(), ref.K8s(), route)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get route")
	}
	if !meta.IsStatusConditionTrue(route.GetConditions(), condition.ConditionTypeReady) {
		return nil, errors.Errorf("route %s is not ready", ref.String())
	}
	return route, nil
}
