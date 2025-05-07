package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	cs "github.com/telekom/controlplane-mono/common-server/pkg/server"
	"github.com/telekom/controlplane-mono/common-server/pkg/server/serve"
	"github.com/telekom/controlplane-mono/secret-manager/api/util"
	"github.com/telekom/controlplane-mono/secret-manager/cmd/server/config"
	"github.com/telekom/controlplane-mono/secret-manager/internal/api"
	"github.com/telekom/controlplane-mono/secret-manager/internal/handler"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend/cache"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend/conjur"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend/kubernetes"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/controller"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrlr "sigs.k8s.io/controller-runtime"
)

const (
	trueStr = "true"
)

var (
	logLevel    string
	disableTls  bool
	tlsCert     string
	tlsKey      string
	address     string
	configFile  string
	backendType string
)

func init() {
	flag.StringVar(&logLevel, "loglevel", "info", "log level")
	flag.BoolVar(&disableTls, "disable-tls", false, "disable TLS")
	flag.StringVar(&tlsCert, "tls-cert", "/etc/tls/tls.crt", "path to TLS certificate")
	flag.StringVar(&tlsKey, "tls-key", "/etc/tls/tls.key", "path to TLS key")
	flag.StringVar(&address, "address", ":8443", "server address")
	flag.StringVar(&configFile, "configfile", "", "path to config file")
	flag.StringVar(&backendType, "backend", "", "backend type (kubernetes, conjur, encrypt)")
}

func setupLog(logLevel string) logr.Logger {
	logCfg := zap.NewProductionConfig()
	logCfg.DisableCaller = true
	logCfg.DisableStacktrace = true
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logCfg.EncoderConfig.TimeKey = "time"
	zapLogLevel, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		zapLogLevel = zapcore.InfoLevel
	}

	logCfg.Level.SetLevel(zapLogLevel)
	zapLog := zap.Must(logCfg.Build())
	return zapr.NewLogger(zapLog)
}

func newController(ctx context.Context, cfg *config.ServerConfig) (c controller.Controller, err error) {
	if backendType != "" {
		cfg.Backend.Type = backendType
	}
	if cfg.Backend.Type == "" {
		cfg.Backend.Type = "kubernetes"
	}
	cacheDuration, err := time.ParseDuration(cfg.Backend.GetDefault("cache_duration", "10s"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cache duration")
	}

	switch cfg.Backend.Type {
	case "conjur":
		conjurWriteApi := conjur.NewWriteApiOrDie()
		conjurReadApi := conjur.NewReadOnlyApiOrDie()

		backend := conjur.NewBackend(conjurWriteApi, conjurReadApi)
		if cfg.Backend.GetDefault("disable_cache", "false") == trueStr {
			backend = cache.NewCachedBackend(backend, cacheDuration)
		}
		onboarder := conjur.NewOnboarder(conjurWriteApi, backend)
		bouncer := conjur.NewDefaultBouncer()
		bouncer.StartN(ctx, 1)
		onboarder.WithBouncer(bouncer)
		c = controller.NewController(backend, onboarder)

	case "kubernetes":
		k8sClient, err := kubernetes.NewCachedClient(ctx, ctrlr.GetConfigOrDie())
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kubernetes client")
		}
		backend := kubernetes.NewBackend(k8sClient)
		if cfg.Backend.GetDefault("disable_cache", "false") == trueStr {
			backend = cache.NewCachedBackend(backend, cacheDuration)
		}
		onboarder := kubernetes.NewOnboarder(k8sClient)
		c = controller.NewController(backend, onboarder)

	default:
		return nil, errors.Errorf("unknown backend type: %s", cfg.Backend.Type)
	}

	return c, nil
}

func main() {
	flag.Parse()
	log := setupLog(logLevel)

	ctx := cs.SignalHandler(context.Background())

	ctrlr.SetLogger(log)
	cfg := config.GetConfigOrDie(configFile)

	ctrl, err := newController(ctx, cfg)
	if err != nil {
		log.Error(err, "failed to create controller")
		return
	}

	appCfg := cs.NewAppConfig()
	appCfg.CtxLog = &log
	appCfg.ErrorHandler = handler.ErrorHandler
	app := cs.NewAppWithConfig(appCfg)

	probesCtrl := cs.NewProbesController()
	probesCtrl.Register(app, cs.ControllerOpts{})

	apiGroup := app.Group("/api")
	handler := api.NewStrictHandler(handler.NewHandler(ctrl), nil)

	if cfg.Security.Enabled {
		opts := []middleware.KubernetesAuthOption{
			middleware.WithTrustedIssuers(cfg.Security.TrustedIssuers...),
			middleware.WithJWKSetURLs(cfg.Security.JWKSetURLs...),
			middleware.WithAccessConfig(cfg.Security.AccessConfig...),
		}
		if util.IsRunningInCluster() {
			log.Info("🔑 Running in cluster")
			opts = append(opts, middleware.WithInClusterIssuer())
		}
		apiGroup.Use(middleware.NewKubernetesAuthz(opts...))
	}

	api.RegisterHandlersWithOptions(apiGroup, handler, api.FiberServerOptions{})

	go func() {
		if disableTls {
			fmt.Println("⚠️\tUsing HTTP instead of HTTPS. This is not secure.")
			if err := app.Listen(address); err != nil {
				log.Error(err, "failed to start server")
				os.Exit(1)
			}
			return
		}

		ctx = logr.NewContext(ctx, log.WithName("server"))
		if err := serve.ServeTLS(ctx, app, address, tlsCert, tlsKey); err != nil {
			log.Error(err, "failed to start server")
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Error(err, "failed to shutdown server")
	}
}
