package feature

import (
	"context"

	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"github.com/telekom/controlplane-mono/gateway/internal/features"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client/plugin"
)

var _ features.Feature = &CustomScopesFeature{}

type CustomScopesFeature struct {
	priority int
}

var InstanceCustomScopesFeature = &CustomScopesFeature{
	priority: InstanceLastMileSecurityFeature.priority - 1,
}

func (f *CustomScopesFeature) Name() gatewayv1.FeatureType {
	return gatewayv1.FeatureTypeCustomScopes
}

func (f *CustomScopesFeature) Priority() int {
	return f.priority
}

func (f *CustomScopesFeature) IsUsed(ctx context.Context, builder features.FeaturesBuilder) bool {
	return !builder.GetRoute().IsProxy() && !builder.GetRoute().Spec.PassThrough
}

func (f *CustomScopesFeature) Apply(ctx context.Context, builder features.FeaturesBuilder) (err error) {
	jumperConfig := builder.JumperConfig()

	if len(jumperConfig.OAuth) > 0 {
		// already populated by external_idp feature
		return nil
	}

	for _, consumer := range builder.GetAllowedConsumers() { // TODO: implement
		jumperConfig.OAuth[plugin.ConsumerId(consumer.Spec.ConsumerName)] = plugin.OauthCredentials{
			Scopes: "custom_scope",
		}
	}

	return nil
}
