package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Config struct {
		URI      string `envconfig:"MONGODB_URI"`
		Database string `envconfig:"MONGODB_DATABASE"`
	}

	MongoDBClient struct {
		*mongo.Client
		databaseName string
	}
)

func NewMongoDBClient(ctx context.Context, config *Config) (*MongoDBClient, error) {
	clientOptions := options.Client().ApplyURI(config.URI)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return &MongoDBClient{Client: client, databaseName: config.Database}, nil
}

func (mc *MongoDBClient) Database() *mongo.Database {
	return mc.Client.Database(mc.databaseName)
}

func createIndex(ctx context.Context, collection *mongo.Collection, keys interface{}, options *options.IndexOptions) error {
	indexModel := mongo.IndexModel{Keys: keys, Options: options}
	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
