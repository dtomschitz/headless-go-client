package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const configCollection = "configs"

type ConfigRepository struct {
	collection *mongo.Collection
}

func NewConfigRepository(ctx context.Context, database *mongo.Database) (*ConfigRepository, error) {
	if err := createIndex(ctx, database.Collection(configCollection), bson.D{{Key: "version", Value: 1}}, options.Index().SetUnique(true)); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &ConfigRepository{
		collection: database.Collection(configCollection),
	}, nil
}

func (r *ConfigRepository) Create(ctx context.Context, config *internal.Config) error {
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return internal.NewConflictError(fmt.Errorf("config with version %s already exists", config.Version))
		}
		return err
	}
	return nil
}

func (r *ConfigRepository) GetByVersion(ctx context.Context, version string) (*internal.Config, error) {
	filter := bson.M{"version": version}
	return r.findOne(ctx, filter)
}

func (r *ConfigRepository) GetLatest(ctx context.Context) (*internal.Config, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})
	return r.findOne(ctx, bson.M{}, opts)
}

func (r *ConfigRepository) findOne(ctx context.Context, filter bson.M, opts ...*options.FindOneOptions) (*internal.Config, error) {
	var config internal.Config
	err := r.collection.FindOne(ctx, filter, opts...).Decode(&config)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, internal.NewNotFoundError(fmt.Errorf("config not found: %w", err))
		}
		return nil, err
	}
	return &config, nil
}
