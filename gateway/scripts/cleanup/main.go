package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	kong "github.com/telekom/controlplane-mono/gateway/pkg/kong/api"
	"github.com/telekom/controlplane-mono/gateway/pkg/kongutil"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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

	kongClient, err := kongutil.NewClientFor(gateway)
	if err != nil {
		panic(err)
	}

	err = Cleanup(ctx, kongClient)
	if err != nil {
		panic(err)
	}
}

func Cleanup(ctx context.Context,
	kongClient kong.ClientWithResponsesInterface,
) error {

	routes, err := kongClient.ListRouteWithResponse(ctx, &kong.ListRouteParams{})
	if err != nil {
		return err
	}
	if routes.StatusCode() != 200 {
		return fmt.Errorf("failed to list routes: %d", routes.StatusCode())
	}

	for _, route := range *routes.JSON200.Data {
		routeName := *route.Name
		if strings.Contains(routeName, "admin") {
			continue
		}

		err = DeleteRouteAndService(ctx, kongClient, route)
		if err != nil {
			return err
		}
		fmt.Printf("🧹 Route %s deleted from Kong\n", routeName)
	}

	return nil
}

func DeleteRouteAndService(ctx context.Context, kongClient kong.ClientWithResponsesInterface, route kong.Route) error {
	_, err := kongClient.DeleteServiceWithResponse(ctx, *route.Service.Id)
	if err != nil {
		return errors.Wrapf(err, "failed to delete service %s", *route.Service.Id)
	}

	_, err = kongClient.DeleteRouteWithResponse(ctx, *route.Id)
	if err != nil {
		return errors.Wrapf(err, "failed to delete route %s", *route.Id)
	}
	return nil
}
