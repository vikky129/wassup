package coll

import (
	"context"
	"encoding/base64"
	"fmt"
	"personal/wassup/spec"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	MessageCollection   = "messages"
	ImageMessageType    = "image"
	AudioMessageType    = "audio"
	VideoMessageType    = "video"
	DocumentMessageType = "document"
	DefaultLimit        = 20
)

type MsgHandler struct {
	collection     *mongo.Collection
	dbName         string
	collectionName string
}

func NewMsgHandler(client *mongo.Client, dbName string) *MsgHandler {
	collection := client.Database(dbName).Collection(MessageCollection)
	return &MsgHandler{
		collection:     collection,
		dbName:         dbName,
		collectionName: MessageCollection,
	}
}

func (mh *MsgHandler) AddMessage(ctx context.Context, msg *spec.Message) error {
	_, err := mh.collection.InsertOne(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func (mh *MsgHandler) GetMessages(ctx context.Context, convID string, limit int, cursor string) (*spec.GetMessagesResponse, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	query := bson.M{"conversation_id": convID}
	findOptions := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit))
	if cursor != "" {
		base64Cursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, err
		}

		cursorInt64, err := strconv.ParseInt(string(base64Cursor), 10, 64)
		if err != nil {
			return nil, err
		}

		query["created_at"] = bson.M{"$lt": cursorInt64}
	}
	mongoCursor, err := mh.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, err
	}
	defer mongoCursor.Close(ctx)

	var messages []spec.Message
	for mongoCursor.Next(ctx) {
		var msg spec.Message
		if err := mongoCursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := mongoCursor.Err(); err != nil {
		return nil, err
	}

	var base64Cursor string
	if len(messages) == limit {
		newCursor := messages[len(messages)-1].CreatedAt
		base64Cursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", newCursor)))
	}

	return &spec.GetMessagesResponse{
		Messages:   messages,
		NextCursor: base64Cursor,
	}, nil
}

func (mh *MsgHandler) UpdateStatusDelivered(ctx context.Context, convIDs []string, userID string) error {
	_, err := mh.collection.UpdateMany(ctx, bson.M{"conversation_id": bson.M{"$in": convIDs}, "sender_id": bson.M{"$ne": userID}, "status": spec.StatusSent}, bson.M{"$set": bson.M{"status": spec.StatusDelivered}})
	return err
}
