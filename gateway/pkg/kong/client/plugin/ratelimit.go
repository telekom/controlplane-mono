package plugin

import (
	"encoding/json"

	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client"
)

var _ client.CustomPlugin = &RateLimitPlugin{}

type Policy string

const (
	PolicyLocal   Policy = "local"
	PolicyCluster Policy = "cluster"
	PolicyRedis   Policy = "redis"
)

type RedisConfig struct {
	Host       string `json:"redis_host,omitempty"`
	Port       int    `json:"redis_port,omitempty"`
	Timeout    int    `json:"redis_timeout,omitempty"`
	Username   string `json:"redis_username,omitempty"`
	Password   string `json:"redis_password,omitempty"`
	Database   int    `json:"redis_database,omitempty"`
	Ssl        bool   `json:"redis_ssl,omitempty"`
	SslVerify  bool   `json:"redis_ssl_verify,omitempty"`
	ServerName string `json:"redis_server_name,omitempty"`
}

type LimitConfig struct {
	Second int `json:"second,omitempty"`
	Minute int `json:"minute,omitempty"`
	Hour   int `json:"hour,omitempty"`
	// add more fields if needed
}

type Limits struct {
	Consumer *LimitConfig `json:"consumer,omitempty"`
	Service  *LimitConfig `json:"service,omitempty"`
}

// See https://docs.konghq.com/hub/kong-inc/rate-limiting/configuration/
type RateLimitPluginConfig struct {
	Policy        Policy `json:"policy,omitempty"`
	FaultTolerant bool   `json:"fault_tolerant,omitempty"`
	RedisConfig   `json:",inline"`

	HideClientHeaders bool   `json:"hide_client_headers,omitempty"`
	ErrorCode         int    `json:"error_code,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`

	OmitConsumer string `json:"omit_consumer,omitempty"` // Custom field

	Limits Limits `json:"limits"` // Custom field
}

type RateLimitPlugin struct {
	Id     string                `json:"id,omitempty"`
	Config RateLimitPluginConfig `json:"config,omitempty"`
	route  *gatewayv1.Route
}

func (p *RateLimitPlugin) GetId() string {
	return p.Id
}

func (p *RateLimitPlugin) SetId(id string) {
	p.Id = id
	p.route.SetProperty("kongRateLimitingPluginId", id)
}

func (p *RateLimitPlugin) GetName() string {
	return "rate-limiting-merged"
}

func (p *RateLimitPlugin) GetRoute() *string {
	return &p.route.Name
}

func (p *RateLimitPlugin) GetConsumer() *string {
	return nil
}

func (p *RateLimitPlugin) GetConfig() map[string]interface{} {
	m := make(map[string]interface{})
	if err := deepCopy(p.Config, &m); err != nil {
		panic(err)
	}
	return m
}

func RateLimitPluginFromRoute(route *gatewayv1.Route) *RateLimitPlugin {
	return &RateLimitPlugin{
		Id: route.GetProperty("kongRateLimitingPluginId"),
		Config: RateLimitPluginConfig{
			FaultTolerant:     true,
			HideClientHeaders: false,
		},
		route: route,
	}
}

func deepCopy[T any](v any, t T) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &t)
}
