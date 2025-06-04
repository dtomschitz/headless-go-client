package main

import (
	"context"
	"log"
	"os"
	"time"

	"log/slog"

	"github.com/dtomschitz/headless-go-client/config"
	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/lifecycle"
	"github.com/dtomschitz/headless-go-client/updater"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	closer, err := lifecycle.NewLifecycleService(ctx, lifecycle.WithLogger(logger))
	if err != nil {
		log.Fatalf("Failed to create lifecycle service: %v", err)
	}
	defer closer.CloseAll(ctx)

	configService, err := config.NewConfigService(ctx, "http://localhost:8080/config", config.WithLogger(logger))
	if err != nil {
		log.Fatalf("Failed to create config service: %v", err)
	}
	closer.Register(configService)

	eventService, err := event.NewService(ctx, "http://localhost:8080/events", time.Hour*1, event.WithLogger(logger))
	if err != nil {
		log.Fatalf("Failed to create event service: %v", err)
	}
	closer.Register(eventService)

	selfUpdater, err := updater.Start(ctx, "dev", updater.WithLogger(logger))
	if err != nil {
		log.Fatalf("Failed to run self updater: %v", err)
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
}
