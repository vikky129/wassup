package coll

import (
	"context"
	"errors"
	"fmt"
	"personal/wassup/cerror"
	"personal/wassup/spec"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const UserCollection = "users"

type UserHandler struct {
	collection     *mongo.Collection
	dbName         string
	collectionName string
}

func NewUserHandler(client *mongo.Client, dbName string) *UserHandler {
	collection := client.Database(dbName).Collection(UserCollection)
	return &UserHandler{
		collection:     collection,
		dbName:         dbName,
		collectionName: UserCollection,
	}
}

func (uh *UserHandler) CreateUser(ctx context.Context, user *spec.User) (string, error) {
	if user.PhoneNumber == "" {
		return "", errors.New(cerror.ErrInvalidPhoneNumber)
	}

	user.ID = bson.NewObjectID().Hex()

	_, err := uh.collection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

func (uh *UserHandler) GetUserByID(ctx context.Context, userID string) (*spec.User, error) {
	if userID == "" {
		return nil, errors.New(cerror.ErrInvalidUserID)
	}

	var user spec.User
	err := uh.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(cerror.ErrUserNotFound)
		}
		return nil, fmt.Errorf("error fetching user by ID: %w", err)
	}

	return &user, nil
}

func (uh *UserHandler) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*spec.User, error) {
	if phoneNumber == "" {
		return nil, errors.New(cerror.ErrInvalidPhoneNumber)
	}

	var user spec.User
	err := uh.collection.FindOne(ctx, map[string]interface{}{"phone_number": phoneNumber}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(cerror.ErrUserNotFound)
		}
		return nil, fmt.Errorf("error fetching user by phone number: %w", err)
	}

	return &user, nil
}

func (uh *UserHandler) UpdateUser(ctx context.Context, user *spec.User) error {
	if user.ID == "" {
		return errors.New(cerror.ErrInvalidUserID)
	}

	_, err := uh.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (uh *UserHandler) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New(cerror.ErrInvalidUserID)
	}

	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New(cerror.ErrInvalidUserID)
	}

	_, err = uh.collection.DeleteOne(ctx, map[string]interface{}{"_id": objID})
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}
