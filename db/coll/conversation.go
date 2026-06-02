package coll

import (
	"context"
	"personal/wassup/spec"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	ConversationCollection = "conversations"
	Admin                  = "admin"
	Member                 = "member"
)

type ConvHandler struct {
	collection     *mongo.Collection
	dbName         string
	collectionName string
}

func NewConvHandler(client *mongo.Client, dbName string) *ConvHandler {
	collection := client.Database(dbName).Collection(ConversationCollection)
	return &ConvHandler{
		collection:     collection,
		dbName:         dbName,
		collectionName: ConversationCollection,
	}
}

func (ch *ConvHandler) CreateConversation(ctx context.Context, conv *spec.Conversation) (string, error) {
	id := uuid.New().String()
	conv.ID = id

	_, err := ch.collection.InsertOne(ctx, conv)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ch *ConvHandler) GetConversationParticipants(ctx context.Context, convID string) ([]spec.Participant, error) {
	var conv spec.Conversation
	err := ch.collection.FindOne(ctx, bson.M{"_id": convID}, options.FindOne().SetProjection(bson.M{"participants": 1})).Decode(&conv)
	if err != nil {
		return nil, err
	}

	return conv.Participants, nil
}

func (ch *ConvHandler) UpdateConversationParticipants(ctx context.Context, convID string, participants []spec.Participant) error {
	_, err := ch.collection.UpdateOne(ctx, bson.M{"_id": convID}, bson.M{"$set": bson.M{"participants": participants, "updated_at": time.Now().Unix()}})
	return err
}

func (ch *ConvHandler) UpdateAddMessage(ctx context.Context, convID string, participants []spec.Participant, updatedAt int64) error {
	_, err := ch.collection.UpdateOne(ctx, bson.M{"_id": convID}, bson.M{"$set": bson.M{"participants": participants, "updated_at": updatedAt, "last_message_at": updatedAt}})
	return err
}

func (ch *ConvHandler) ListConversations(ctx context.Context, userID string) ([]spec.Conversation, error) {
	result, err := ch.collection.Find(ctx, bson.M{"participants.user_id": userID}, options.Find().SetSort(bson.M{"last_message_at": -1}))
	if err != nil {
		return nil, err
	}
	defer result.Close(ctx)

	var conversations []spec.Conversation
	for result.Next(ctx) {
		var conv spec.Conversation
		if err := result.Decode(&conv); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (ch *ConvHandler) GetUnReadConversations(ctx context.Context, userID string) ([]string, error) {
	result, err := ch.collection.Find(ctx, bson.M{"participants.user_id": userID, "participants.unread_message_count": bson.M{"$gt": 0}}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer result.Close(ctx)

	var convIDs []string
	for result.Next(ctx) {
		var conv struct {
			ID string `bson:"_id"`
		}
		if err := result.Decode(&conv); err != nil {
			return nil, err
		}
		convIDs = append(convIDs, conv.ID)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return convIDs, nil
}

func (ch *ConvHandler) UpdateConversationParticipantStatus(ctx context.Context, convID, userID, status string) error {
	filter := bson.M{"_id": convID, "participants.user_id": userID}
	update := bson.M{"$set": bson.M{"participants.$.message_status": status}}
	_, err := ch.collection.UpdateOne(ctx, filter, update)
	return err
}
