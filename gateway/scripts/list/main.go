package main

import (
	"context"
	"encoding/json"
	"strings"

	kong "github.com/telekom/controlplane-mono/gateway/pkg/kong/api"
	"github.com/telekom/controlplane-mono/gateway/pkg/kongutil"
	"k8s.io/utils/ptr"
)

var rootCtx = context.Background()

func main() {

	gwCfg := kongutil.NewGatewayConfig("https://stargate-ce-admin-integration.test.dhei.telekom.de/admin-api", "rover", "<secret>", "https://iris-integration.test.dhei.telekom.de/auth/realms/rover")

	kc, err := kongutil.NewClientFor(gwCfg)
	if err != nil {
		panic(err)
	}

	res, err := kc.ListPluginWithResponse(rootCtx, &kong.ListPluginParams{
		Tags: ptr.To(strings.Join([]string{"rate-limiting-merged"}, ",")),
	})

	if err != nil {
		panic(err)
	}

	type resBody struct {
		Data []kong.Plugin `json:"Data"`
	}

	var body resBody
	if err := json.Unmarshal(res.Body, &body); err != nil {
		panic(err)
	}

	for _, p := range body.Data {
		b, _ := json.MarshalIndent(p.Config, "", "  ")
		println(string(b))
		println(strings.Repeat("-", 80))
	}

}
