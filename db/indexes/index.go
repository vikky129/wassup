package indexes

import (
	"context"
	"fmt"
	"personal/wassup/db/coll"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Version int = 3
var CollName = "indexes"

type IndexCollection struct {
	Version   int   `bson:"version"`
	CreatedAt int64 `bson:"created_at"`
	UpdatedAt int64 `bson:"updated_at"`
}

type IndexHandler struct {
	client *mongo.Client
}

func NewIndexHandler(client *mongo.Client) *IndexHandler {
	return &IndexHandler{
		client: client,
	}
}

func (ih *IndexHandler) indexVersionOne(ctx context.Context, database string) error {
	// Create unique index on users collection for phone_number
	ih.client.Database(database).Collection(coll.UserCollection).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "phone_number", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	ih.client.Database(database).Collection(coll.MessageCollection).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "conversation_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	)

	ih.client.Database(database).Collection(coll.ConversationCollection).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "participants.user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	)

	return nil
}

func (ih *IndexHandler) indexVersionTwo(ctx context.Context, database string) error {
	ih.client.Database(database).Collection(coll.ConversationCollection).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "participants.user_id", Value: 1},
				{Key: "last_message_at", Value: -1},
			},
		},
	)

	return nil
}

func (ih *IndexHandler) indexVersionThree(ctx context.Context, database string) error {
	_, err := ih.client.Database(database).Collection(coll.MessageCollection).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "conversation_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	)

	return err
}

func (ih *IndexHandler) AddIndexes(ctx context.Context, database string) error {
	var indexColl IndexCollection
	err := ih.client.Database(database).Collection(CollName).FindOne(ctx, bson.D{}).Decode(&indexColl)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			indexColl = IndexCollection{
				Version:   0,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}

			_, err = ih.client.Database(database).Collection(CollName).InsertOne(ctx, indexColl)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("error %w occured fetching index collection", err)
		}
	}

	if indexColl.Version == 0 {
		err = ih.client.Database(database).Collection(CollName).FindOne(ctx, bson.D{}).Decode(&indexColl)
		if err != nil {
			return fmt.Errorf("error %w occured fetching index collection", err)
		}
	}

	var currVersion int = indexColl.Version + 1
	for currVersion <= Version {
		switch currVersion {
		case 1:
			err := ih.indexVersionOne(ctx, database)
			if err != nil {
				return err
			}
			currVersion++
		case 2:
			err := ih.indexVersionTwo(ctx, database)
			if err != nil {
				return err
			}
			currVersion++
		case 3:
			err := ih.indexVersionThree(ctx, database)
			if err != nil {
				return err
			}
			currVersion++
		default:
			return fmt.Errorf("no migration found for version %d", currVersion)
		}

		if currVersion > Version {
			indexColl.Version = Version
			if indexColl.CreatedAt == 0 {
				indexColl.CreatedAt = time.Now().Unix()
			}

			indexColl.UpdatedAt = time.Now().Unix()
			_, err = ih.client.Database(database).Collection(CollName).UpdateOne(ctx, bson.D{}, bson.D{{Key: "$set", Value: indexColl}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
