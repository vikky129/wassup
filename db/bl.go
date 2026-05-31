package db

import (
	"context"
	"fmt"
	"personal/wassup/config"
	"personal/wassup/db/coll"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBHandler struct {
	client         *mongo.Client
	UserHandler    *coll.UserHandler
	ConvHandler    *coll.ConvHandler
	MessageHandler *coll.MsgHandler
	config         *config.Config
}

func NewDBHandler(config *config.Config) *DBHandler {
	return &DBHandler{config: config}
}

func (dbh *DBHandler) GetClient() (*mongo.Client, error) {
	if dbh.client == nil {
		return nil, fmt.Errorf("mongo client is not initialized")
	}
	return dbh.client, nil
}

func (dbh *DBHandler) Init() error {
	client, err := mongo.Connect(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:27017", dbh.config.MongoContainerName)))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("failed to ping mongo: %w", err)
	}

	dbh.client = client
	dbh.UserHandler = coll.NewUserHandler(client, DBName)
	dbh.ConvHandler = coll.NewConvHandler(client, DBName)
	dbh.MessageHandler = coll.NewMsgHandler(client, DBName)
	return nil
}
