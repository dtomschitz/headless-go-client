package main

import (
	"context"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"github.com/dtomschitz/headless-go-client/example/backend/internal/database"
	"github.com/dtomschitz/headless-go-client/example/backend/internal/http"
	"github.com/kelseyhightower/envconfig"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	logger := internal.NewLogger(ctx)

	var databaseConfig database.Config
	if err := envconfig.Process("", &databaseConfig); err != nil {
		logger.Fatalf("failed to process database config: %v", err)
	}

	databaseClient, err := database.NewMongoDBClient(ctx, &databaseConfig)
	if err != nil {
		logger.Fatalf("failed to create MongoDB client: %v", err)
	}

	configRepository, err := database.NewConfigRepository(ctx, databaseClient.Database())
	if err != nil {
		logger.Fatalf("failed to create config repository: %v", err)
	}

	configService := internal.NewConfigService(configRepository)

	if err := http.StartServer(configService); err != nil {
		logger.Fatalf("failed to start HTTP server: %v", err)
	}
}
