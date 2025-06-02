package main

import (
	"context"
	"github.com/dtomschitz/headless-go-client/updater"
	"log"
)

func main() {
	ctx := context.Background()

	selfUpdater, err := updater.Start(ctx, "dev")
	if err != nil {
		log.Fatalf("Failed to run self updater: %v", err)
	}
	selfUpdater.ListenForUpdateAvailable(ctx, func(ctx context.Context, manifest *updater.Manifest) {
		err := selfUpdater.ApplyUpdate(ctx, manifest)
		if err != nil {
			return
		}
	})
	selfUpdater.ListenForUpdateApplied(ctx, func(ctx context.Context, manifest *updater.Manifest) {

	})
}
