package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
	"github.com/dtomschitz/headless-go-client/config"
	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/lifecycle"
	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/dtomschitz/headless-go-client/manifest"
	"github.com/dtomschitz/headless-go-client/updater"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ConfigManifestURL string `envconfig:"CONFIG_MANIFEST_URL" required:"true"`
	ConfigStorageURL  string `envconfig:"CONFIG_STORAGE_URL" required:"true" default:"./config.json"`
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, commonCtx.DeviceIdKey, "12345")
	ctx = context.WithValue(ctx, commonCtx.ClientVersionKey, "dev")

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.SlogFactory(ctx)
	log.Info("starting client")

	clientConfig := Config{}
	err := envconfig.Process("", &clientConfig)
	if err != nil {
		log.Error("failed to process client config", err)
		return
	}

	closer, err := lifecycle.NewService(ctx, lifecycle.WithLogger(logger.SlogFactory))
	if err != nil {
		log.Error("failed to create lifecycle service", err)
		return
	}
	defer closer.CloseAll(ctx)

	configStorage := config.NewFileStorage(clientConfig.ConfigStorageURL)
	configService, err := config.NewService(ctx, clientConfig.ConfigManifestURL, config.WithLogger(logger.SlogFactory), config.WithStorage(configStorage))
	if err != nil {
		log.Error("failed to create config service", err)
		return
	}
	closer.Register(configService)

	eventService, err := event.NewService(ctx, "http://localhost:8080/events", event.WithLogger(logger.SlogFactory))
	if err != nil {
		log.Error("failed to create event service", err)
		return
	}
	closer.Register(eventService)

	selfUpdater, err := updater.NewService(ctx, "dev", updater.WithLogger(logger.SlogFactory))
	if err != nil {
		log.Error("failed to create update service", err)
		return
	}
	eventService.RegisterProducer(selfUpdater)
	closer.Register(selfUpdater)

	selfUpdater.ListenForUpdateAvailable(ctx, func(ctx context.Context, manifest *manifest.Manifest) {
		err := selfUpdater.ApplyUpdate(ctx, manifest)
		if err != nil {

			return
		}
	})
	selfUpdater.ListenForUpdateApplied(ctx, func(ctx context.Context, manifest *manifest.Manifest) {

	})

	<-ctx.Done()
	log.Info("client is going to shutdown")
}
