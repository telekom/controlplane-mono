package feature

import (
	"context"
	"strconv"
	"strings"

	"github.com/telekom/controlplane-mono/common/pkg/util/contextutil"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client/plugin"

	"github.com/telekom/controlplane-mono/gateway/internal/features"
)

var _ features.Feature = &LastMileSecurityFeature{}

type LastMileSecurityFeature struct {
	priority int
}

var InstanceLastMileSecurityFeature = &LastMileSecurityFeature{
	priority: 100,
}

func (f *LastMileSecurityFeature) Name() gatewayv1.FeatureType {
	return gatewayv1.FeatureTypeLastMileSecurity
}

func (f *LastMileSecurityFeature) Priority() int {
	return f.priority
}

func (f *LastMileSecurityFeature) IsUsed(ctx context.Context, builder features.FeaturesBuilder) bool {
	route := builder.GetRoute()
	return !route.Spec.PassThrough
}

func (f *LastMileSecurityFeature) Apply(ctx context.Context, builder features.FeaturesBuilder) (err error) {
	route := builder.GetRoute()
	realm := builder.GetRealm()
	envName := contextutil.EnvFromContextOrDie(ctx)

	rtpPlugin := builder.RequestTransformerPlugin()

	builder.SetUpstream(client.NewUpstreamOrDie("http://localhost:8080/proxy"))

	if route.IsProxy() {
		// Proxy Route

		rtpPlugin.Config.Append.
			AddHeader("issuer", route.Spec.Upstreams[0].IssuerUrl).
			AddHeader("client_id", route.Spec.Upstreams[0].ClientId).
			AddHeader("client_secret", route.Spec.Upstreams[0].ClientSecret).
			AddHeader("remote_api_url", CreateRemoteApiUrl(route)).
			AddHeader(plugin.JumperConfigKey, plugin.ToBase64OrDie(builder.JumperConfig()))

	} else {
		// Real Route

		rtpPlugin.Config.Remove.AddHeader("consumer-token")

		rtpPlugin.Config.Replace.
			AddHeader("Authorization", "$(headers['consumer-token'] or headers['Authorization'])")

		rtpPlugin.Config.Append.
			AddHeader("remote_api_url", CreateRemoteApiUrl(route)).
			AddHeader("api_base_path", route.Spec.Upstreams[0].Path).
			AddHeader("access_token_forwarding", "false").
			AddHeader(plugin.JumperConfigKey, plugin.ToBase64OrDie(builder.JumperConfig()))

		// We could use append here but then in a cross-CP mesh scenario we would have multiple headers like "realm1,realm2"
		// Add them if they are not present yet
		rtpPlugin.Config.Add.
			AddHeader("environment", envName).
			AddHeader("realm", realm.Name)
		// Ensure that we replace any existing headers in case they were already set
		rtpPlugin.Config.Replace.
			AddHeader("environment", envName).
			AddHeader("realm", realm.Name)
	}

	return nil
}

func CreateRemoteApiUrl(route *gatewayv1.Route) string {
	upstream := route.Spec.Upstreams[0]

	result := upstream.Host
	if upstream.Port != 0 {
		result = result + ":" + strconv.Itoa(upstream.Port)
	}
	result = result + upstream.Path

	result = strings.ReplaceAll(result, "//", "/")

	return upstream.Scheme + "://" + result
}
