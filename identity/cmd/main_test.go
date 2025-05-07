package main

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func TestMetricsServerIsConfiguredCorrectly(t *testing.T) {
	var metricsAddr string
	var secureMetrics bool
	var tlsOpts []func(*tls.Config)

	metricsAddr = ":8443"
	secureMetrics = true

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	assert.Equal(t, ":8443", metricsServerOptions.BindAddress)
	assert.True(t, metricsServerOptions.SecureServing)
}

func TestWebhookServerIsConfiguredCorrectly(t *testing.T) {
	var tlsOpts []func(*tls.Config)
	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	assert.NotNil(t, webhookServer)
}

func TestManagerIsStartedSuccessfully(t *testing.T) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{},
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: ":8081",
		LeaderElection:         false,
		LeaderElectionID:       "f09b7a29.cp.ei.telekom.de",
	})

	assert.NoError(t, err)
	assert.NotNil(t, mgr)
}
