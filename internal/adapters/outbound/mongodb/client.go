// Package mongodb provides MongoDB adapter implementations.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// MongoConfig holds MongoDB connection configuration.
type MongoConfig struct {
	URI      string
	Database string
	MinPool  uint64
	MaxPool  uint64
	Timeout  time.Duration
	AppName  string
}

// NewMongoClient creates and pings a new MongoDB client.
func NewMongoClient(ctx context.Context, cfg MongoConfig, log logger.Logger) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMinPoolSize(cfg.MinPool).
		SetMaxPoolSize(cfg.MaxPool).
		SetServerSelectionTimeout(cfg.Timeout).
		SetConnectTimeout(cfg.Timeout).
		SetTimeout(cfg.Timeout).
		SetAppName(cfg.AppName)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb.NewMongoClient: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("mongodb.NewMongoClient: ping: %w", err)
	}

	log.Info("connected to MongoDB")

	return client, nil
}

// Close disconnects the MongoDB client.
func Close(ctx context.Context, client *mongo.Client, log logger.Logger) {
	if err := client.Disconnect(ctx); err != nil {
		log.Error("failed to disconnect MongoDB", zap.Error(err))
	} else {
		log.Info("disconnected from MongoDB")
	}
}
