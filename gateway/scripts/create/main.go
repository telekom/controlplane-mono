package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/telekom/controlplane-mono/common/pkg/types"
	"github.com/telekom/controlplane-mono/common/pkg/util/contextutil"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	cc "github.com/telekom/controlplane-mono/common/pkg/client"

	"github.com/telekom/controlplane-mono/gateway/internal/features/feature"
	route_handler "github.com/telekom/controlplane-mono/gateway/internal/handler/route"
)

var (
	ctx         = context.Background()
	kubeContext string
	environment string
	zone        string
	gateway     string
)

func init() {
	flag.StringVar(&kubeContext, "context", "", "Kube context")
	flag.StringVar(&environment, "env", "", "Environment")
	flag.StringVar(&zone, "zone", "", "Zone")
	flag.StringVar(&gateway, "gateway", "", "Gateway instance")
}

func NewClientOrDie() client.Client {
	cfg, err := config.GetConfigWithContext(kubeContext)
	if err != nil {
		panic(err)
	}
	utilruntime.Must(gatewayv1.AddToScheme(scheme.Scheme))
	c, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err)
	}

	return c
}

func main() {
	flag.Parse()

	k8sClient := NewClientOrDie()

	gateway := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gateway,
			Namespace: fmt.Sprintf("%s--%s", environment, zone),
		},
	}
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gateway), gateway)
	if err != nil {
		panic(err)
	}

	ctx := contextutil.WithEnv(ctx, "poc")
	ctx = cc.WithClient(ctx, cc.NewJanitorClient(cc.NewScopedClient(k8sClient, "poc")))

	route := &gatewayv1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-namespace",
		},
		Spec: gatewayv1.RouteSpec{
			Realm: types.ObjectRef{
				Name:      "poc",
				Namespace: "poc--" + zone,
			},
			Upstreams: []gatewayv1.Upstream{
				{
					Scheme: "https",
					Host:   "httpbin.org",
					Port:   443,
					Path:   "/anything",
				},
			},
			Downstreams: []gatewayv1.Downstream{
				{
					Host:      "stargate-distcp2-dataplane1.dev.dhei.telekom.de",
					Path:      "/test-route123",
					IssuerUrl: "https://iris-distcp1-dataplane2.dev.dhei.telekom.de/auth/realms/poc",
				},
			},
		},
	}

	builder, err := route_handler.NewFeatureBuilder(ctx, route)
	if err != nil {
		panic(err)
	}

	builder.EnableFeature(feature.InstanceAccessControlFeature)
	builder.EnableFeature(feature.InstancePassThroughFeature)
	builder.EnableFeature(feature.InstanceLastMileSecurityFeature)

	err = builder.Build(ctx)
	if err != nil {
		panic(err)
	}

	printJson(route.Status)

}

func printJson(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))

}
