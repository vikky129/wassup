package proto

import (
	"context"
	"fmt"
	"log"
	"personal/wassup/redis"
	"personal/wassup/spec"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	redisHandler *redis.CacheHandler
}

func NewGrpcClient(redisHandler *redis.CacheHandler) *GrpcClient {
	return &GrpcClient{redisHandler: redisHandler}
}
func (grpcClient *GrpcClient) SendMessage(ctx context.Context, senderID string, participants []string, message *spec.Message) error {
	for _, participant := range participants {
		if participant == senderID {
			continue
		}

		recipientLive, err := grpcClient.redisHandler.CheckKeyPrefixExists(participant)
		if err != nil {
			log.Printf("Error checking live status for user %s: %v", participant, err)
			return err
		}

		log.Printf("User %s live status: %v\n", participant, recipientLive)
		if recipientLive {
			containerAddress, err := grpcClient.redisHandler.GetContainerDetails(ctx, participant)
			if err != nil {
				fmt.Println("Error getting container details:", err)
				continue
			}

			for _, addressMap := range containerAddress {
				for key, address := range addressMap {
					fmt.Printf("Sending message to user %s at address %s\n", participant, address)
					if strings.HasSuffix(key, "name") {
						continue
					}
					go func(address, participant string, message *spec.Message) {
						err := sendMessageViaGrpc(address, participant, message)
						if err != nil {
							fmt.Println("GRPC Error", err)
						}
					}(address, participant, message)
				}
			}
		}
	}

	return nil
}
func sendMessageViaGrpc(address, receiverID string, message *spec.Message) error {
	grpcConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	c := NewMessengerClient(grpcConn)
	_, err = c.SendMessage(context.Background(), &AddMessageRequest{
		ReceiverId: receiverID,
		Message: &AddMessageInfo{
			XId:            message.ID,
			ConversationId: message.ConversationID,
			Text:           message.Text,
			Type:           message.Type,
			SenderId:       message.SenderID,
			CreatedAt:      message.CreatedAt,
			UpdatedAt:      message.UpdatedAt,
			MediaUrl:       message.MediaURL,
		},
	})

	if err != nil {
		fmt.Println("Error sending message via gRPC:", err)
	}
	return err
}

func (grpcClient *GrpcClient) UpdateParticipantStatusForConversation(ctx context.Context, convID, userID string, participants []string, status string) error {
	liveUserAddresses, err := grpcClient.redisHandler.GetLiveUserAddressForUser(context.Background(), userID, participants)
	if err != nil {
		fmt.Println("failed getting live user addresses from cache:", err)
		return err
	}

	for participantID, addresses := range liveUserAddresses {
		for _, address := range addresses {
			fmt.Printf("Sending read receipt to user %s at address %s\n", participantID, address)
			go func(address, participantID, convID string) {
				err := UpdateParticipantStatusForConversationViaGrpc(address, participantID, convID, string(spec.StatusRead))
				if err != nil {
					fmt.Printf("failed sending read receipt to user %s at address %s: %v\n", participantID, address, err)
				}
			}(address, participantID, convID)
		}
		fmt.Printf("Successfully sent read receipt to user %s\n", participantID)
	}
	return nil
}

func UpdateParticipantStatusForConversationViaGrpc(address, userID, convID, status string) error {
	grpcConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	c := NewMessengerClient(grpcConn)
	_, err = c.UpdateParticipantStatusForConversation(context.Background(), &UpdateParticipantStatusRequest{
		UserId:         userID,
		ConversationId: convID,
		Status:         status,
	})

	if err != nil {
		fmt.Println("Error updating participant status via gRPC:", err)
	}
	return err
}
