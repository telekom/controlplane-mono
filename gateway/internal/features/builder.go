package features

import (
	"context"
	"sort"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"

	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client/plugin"
)

type Feature interface {
	// Name of the feature
	Name() gatewayv1.FeatureType
	// Priority of this feature in the feature-chain
	// The higher the priority, the later the feature is applied
	// Feature can have a relative priority to other features
	Priority() int
	// IsUsed checks if the feature is used in the current builder-context
	IsUsed(ctx context.Context, builder FeaturesBuilder) bool
	// Apply applies the feature to the current builder-context
	// It may modify the plugins and upstream of the builder-context
	Apply(ctx context.Context, builder FeaturesBuilder) error
}

//go:generate mockgen -source=builder.go -destination=mock/builder.gen.go -package=mock
type FeaturesBuilder interface {
	EnableFeature(f Feature)
	GetRoute() *gatewayv1.Route
	GetRealm() *gatewayv1.Realm
	GetGateway() *gatewayv1.Gateway
	GetAllowedConsumers() []*gatewayv1.ConsumeRoute
	AddAllowedConsumers(...*gatewayv1.ConsumeRoute)

	SetUpstream(client.Upstream)
	RequestTransformerPlugin() *plugin.RequestTransformerPlugin
	AclPlugin() *plugin.AclPlugin
	JwtPlugin() *plugin.JwtPlugin
	RateLimitPlugin() *plugin.RateLimitPlugin
	JumperConfig() *plugin.JumperConfig

	Build(context.Context) error
}

var _ FeaturesBuilder = &Builder{}

type Builder struct {
	kc client.KongClient

	AllowedConsumers []*gatewayv1.ConsumeRoute
	Route            *gatewayv1.Route
	Realm            *gatewayv1.Realm
	Gateway          *gatewayv1.Gateway

	Upstream     client.Upstream
	Plugins      map[string]client.CustomPlugin
	jumperConfig *plugin.JumperConfig
	Features     map[gatewayv1.FeatureType]Feature
}

var NewFeatureBuilder = func(kc client.KongClient, route *gatewayv1.Route, realm *gatewayv1.Realm, gateway *gatewayv1.Gateway) FeaturesBuilder {
	return &Builder{
		kc: kc,

		AllowedConsumers: []*gatewayv1.ConsumeRoute{},
		Route:            route,
		Realm:            realm,
		Gateway:          gateway,

		Plugins:  map[string]client.CustomPlugin{},
		Features: map[gatewayv1.FeatureType]Feature{},
	}
}

func (b *Builder) EnableFeature(f Feature) {
	b.Features[f.Name()] = f
}

func (b *Builder) GetRoute() *gatewayv1.Route {
	return b.Route
}

func (b *Builder) GetRealm() *gatewayv1.Realm {
	return b.Realm
}

func (b *Builder) GetGateway() *gatewayv1.Gateway {
	return b.Gateway
}

func (b *Builder) GetAllowedConsumers() []*gatewayv1.ConsumeRoute {
	return b.AllowedConsumers
}

func (b *Builder) AddAllowedConsumers(consumers ...*gatewayv1.ConsumeRoute) {
	b.AllowedConsumers = append(b.AllowedConsumers, consumers...)
}

func (b *Builder) RequestTransformerPlugin() *plugin.RequestTransformerPlugin {
	var rtpPlugin *plugin.RequestTransformerPlugin

	if p, ok := b.Plugins["request-transformer"]; ok {
		rtpPlugin, ok = p.(*plugin.RequestTransformerPlugin)
		if !ok {
			panic("plugin is not a RequestTransformerPlugin")
		}
	} else {
		rtpPlugin = plugin.RequestTransformerPluginFromRoute(b.Route)
		b.Plugins["request-transformer"] = rtpPlugin
	}

	return rtpPlugin
}

func (b *Builder) AclPlugin() *plugin.AclPlugin {
	var aclPlugin *plugin.AclPlugin

	if p, ok := b.Plugins["acl"]; ok {
		aclPlugin, ok = p.(*plugin.AclPlugin)
		if !ok {
			panic("plugin is not a AclPlugin")
		}
	} else {
		aclPlugin = plugin.AclPluginFromRoute(b.Route)
		b.Plugins["acl"] = aclPlugin
	}

	return aclPlugin
}

func (b *Builder) JwtPlugin() *plugin.JwtPlugin {
	var jwtPlugin *plugin.JwtPlugin

	if p, ok := b.Plugins["jwt"]; ok {
		jwtPlugin, ok = p.(*plugin.JwtPlugin)
		if !ok {
			panic("plugin is not a JwtPlugin")
		}
	} else {
		jwtPlugin = plugin.JwtPluginFromRoute(b.Route)
		b.Plugins["jwt"] = jwtPlugin
	}

	return jwtPlugin
}

func (b *Builder) RateLimitPlugin() *plugin.RateLimitPlugin {
	var rateLimitPlugin *plugin.RateLimitPlugin

	if p, ok := b.Plugins["rate-limiting"]; ok {
		rateLimitPlugin, ok = p.(*plugin.RateLimitPlugin)
		if !ok {
			panic("plugin is not a RateLimitPlugin")
		}
	} else {
		rateLimitPlugin = plugin.RateLimitPluginFromRoute(b.Route)
		b.Plugins["rate-limiting"] = rateLimitPlugin
	}

	return rateLimitPlugin
}

func (b *Builder) JumperConfig() *plugin.JumperConfig {
	if b.jumperConfig == nil {
		b.jumperConfig = plugin.NewJumperConfig()
	}
	return b.jumperConfig
}

func (b *Builder) SetUpstream(upstream client.Upstream) {
	b.Upstream = upstream
}

func (b *Builder) Build(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx).WithName("features.builder").WithValues("route", b.Route.Name)

	for _, f := range sortFeatures(toSlice(b.Features)) {
		if f.IsUsed(ctx, b) {
			log.V(1).Info("Applying feature", "name", f.Name())
			err := f.Apply(ctx, b)
			if err != nil {
				return err
			}
		} else {
			log.V(1).Info("Feature is not used", "name", f.Name())
		}
	}

	if b.Upstream == nil {
		return errors.New("upstream is not set")
	}

	// In case a plugin was used before but is not used anymore, we need to remove it
	b.Route.Status.Properties = map[string]string{}

	err := b.kc.CreateOrReplaceRoute(ctx, b.Route, b.Upstream)
	if err != nil {
		return errors.Wrap(err, "failed to create or replace route")
	}

	for pn, p := range b.Plugins {
		_, err = b.kc.CreateOrReplacePlugin(ctx, p)
		if err != nil {
			return errors.Wrapf(err, "failed to create or replace plugin %s", pn)
		}
	}

	err = b.kc.CleanupPlugins(ctx, b.Route, toSlice(b.Plugins))
	if err != nil {
		return errors.Wrap(err, "failed to cleanup plugins")
	}

	return nil
}

// sort features based on their priority
// the higher the priority, the later the feature is applied
// this is important because some features might depend on other features
func sortFeatures(featureList []Feature) []Feature {
	sort.Slice(featureList, func(i, j int) bool {
		return featureList[i].Priority() < featureList[j].Priority()
	})
	return featureList
}

func toSlice[K comparable, T any](m map[K]T) []T {
	s := make([]T, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}
