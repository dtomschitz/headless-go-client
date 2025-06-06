package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtomschitz/headless-go-client/config"
	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/lifecycle"
	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/dtomschitz/headless-go-client/updater"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.SlogFactory(ctx)
	log.Info("starting client")

	closer, err := lifecycle.NewService(ctx, lifecycle.WithLogger(logger.SlogFactory))
	if err != nil {
		log.Error("failed to create lifecycle service", err)
		return
	}
	defer closer.CloseAll(ctx)

	configService, err := config.NewService(ctx, "http://localhost:8080/config", config.WithLogger(logger.SlogFactory))
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

	selfUpdater.ListenForUpdateAvailable(ctx, func(ctx context.Context, manifest *updater.Manifest) {
		err := selfUpdater.ApplyUpdate(ctx, manifest)
		if err != nil {

			return
		}
	})
	selfUpdater.ListenForUpdateApplied(ctx, func(ctx context.Context, manifest *updater.Manifest) {

	})

	<-ctx.Done()
	log.Info("client is going to shutdown")
}
