package db

import (
	"context"
	"fmt"
	"personal/wassup/config"
	"personal/wassup/db/coll"
	"personal/wassup/spec"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBHandlerInterface interface {
	GetClient() (*mongo.Client, error)
	Init() error
}

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

func (dbh *DBHandler) Close(ctx context.Context) error {
	if dbh.client == nil {
		return nil
	}
	return dbh.client.Disconnect(ctx)
}

func (dbh *DBHandler) CreateUser(ctx context.Context, user *spec.User) (string, error) {
	return dbh.UserHandler.CreateUser(ctx, user)
}

func (dbh *DBHandler) GetUserByID(ctx context.Context, userID string) (*spec.User, error) {
	return dbh.UserHandler.GetUserByID(ctx, userID)
}

func (dbh *DBHandler) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*spec.User, error) {
	return dbh.UserHandler.GetUserByPhoneNumber(ctx, phoneNumber)
}

func (dbh *DBHandler) UpdateUser(ctx context.Context, user *spec.User) error {
	return dbh.UserHandler.UpdateUser(ctx, user)
}

func (dbh *DBHandler) DeleteUser(ctx context.Context, userID string) error {
	return dbh.UserHandler.DeleteUser(ctx, userID)
}

func (dbh *DBHandler) GetConversationParticipants(ctx context.Context, convID string) ([]spec.Participant, error) {
	return dbh.ConvHandler.GetConversationParticipants(ctx, convID)
}

func (dbh *DBHandler) CreateConversation(ctx context.Context, conv *spec.Conversation) (string, error) {
	return dbh.ConvHandler.CreateConversation(ctx, conv)
}

func (dbh *DBHandler) UpdateConversationParticipants(ctx context.Context, convID string, participants []spec.Participant) error {
	return dbh.ConvHandler.UpdateConversationParticipants(ctx, convID, participants)
}

func (dbh *DBHandler) UpdateAddMessage(ctx context.Context, convID string, participants []spec.Participant, updatedAt int64) error {
	return dbh.ConvHandler.UpdateAddMessage(ctx, convID, participants, updatedAt)
}

func (dbh *DBHandler) ListConversations(ctx context.Context, userID string) ([]spec.Conversation, error) {
	return dbh.ConvHandler.ListConversations(ctx, userID)
}

func (dbh *DBHandler) GetUnReadConversations(ctx context.Context, userID string) ([]string, error) {
	return dbh.ConvHandler.GetUnReadConversations(ctx, userID)
}

func (dbh *DBHandler) UpdateConversationParticipantStatus(ctx context.Context, convID, userID, status string) error {
	return dbh.ConvHandler.UpdateConversationParticipantStatus(ctx, convID, userID, status)
}

func (dbh *DBHandler) AddMessage(ctx context.Context, msg *spec.Message) error {
	return dbh.MessageHandler.AddMessage(ctx, msg)
}

func (dbh *DBHandler) GetMessages(ctx context.Context, convID string, limit int, cursor string) (*spec.GetMessagesResponse, error) {
	return dbh.MessageHandler.GetMessages(ctx, convID, limit, cursor)
}

func (dbh *DBHandler) UpdateStatusDelivered(ctx context.Context, convIDs []string, userID string) error {
	return dbh.MessageHandler.UpdateStatusDelivered(ctx, convIDs, userID)
}

func (dbh *DBHandler) UpdateStatusRead(ctx context.Context, convID string, userID string) error {
	return dbh.MessageHandler.UpdateStatusRead(ctx, convID, userID)
}
