package bl

import (
	"context"
	"fmt"
	"log"
	"personal/wassup/proto"
	"personal/wassup/spec"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MessageTypeText  = "text"
	MessageTypeImage = "image"
	MessageTypeVideo = "video"
	MessageTypeAudio = "audio"
	MessageTypeFile  = "file"
)

func (wh *WassupHandler) AddMessage(ctx context.Context, req *spec.AddMessageRequest) (*spec.AddMessageResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	var timeNow = time.Now().Unix()
	var participants []spec.Participant

	var message = &spec.Message{
		ID:             uuid.NewString(),
		ConversationID: req.ConversationID,
		Text:           req.Text,
		Type:           MessageTypeText,
		SenderID:       userID,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
		Status:         spec.StatusSent,
	}

	if req.IsNewConversation {
		conversation := &spec.Conversation{
			Participants: []spec.Participant{
				{
					UserID:   userID,
					JoinedAt: timeNow,
					Role:     ParticipantTypeMember,
				},
				{
					UserID:   req.ReceiverID,
					JoinedAt: timeNow,
					Role:     ParticipantTypeMember,
				},
			},
		}

		convID, err := wh.dbHandler.ConvHandler.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, err
		}
		message.ConversationID = convID
		participants = conversation.Participants
	}

	err := wh.dbHandler.MessageHandler.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	if len(participants) == 0 {
		participants, err = wh.dbHandler.ConvHandler.GetConversationParticipants(ctx, req.ConversationID)
		if err != nil {
			return nil, err
		}
	}

	participants = incrementUnreadCount(participants, userID)
	err = wh.dbHandler.ConvHandler.UpdateAddMessage(ctx, req.ConversationID, participants, timeNow)
	if err != nil {
		return nil, err
	}

	go wh.LiveSendRecipients(message.ConversationID, userID, participants, message) // Send message to live recipients via gRPC

	return &spec.AddMessageResponse{
		ConversationID: message.ConversationID,
		Message:        message.Text,
	}, nil
}

func (wh *WassupHandler) AddMediaMessage(ctx context.Context, req *spec.AddMediaMessageRequest, filePath string) (*spec.AddMessageResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	var timeNow = time.Now().Unix()
	var message = &spec.Message{
		ID:             uuid.NewString(),
		ConversationID: req.ConversationID,
		Type:           req.MediaType,
		SenderID:       userID,
		MediaURL:       filePath,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
		Status:         spec.StatusSent,
	}

	if req.IsNewConversation {
		conversation := &spec.Conversation{
			Participants: []spec.Participant{
				{
					UserID:   userID,
					Role:     ParticipantTypeMember,
					JoinedAt: timeNow,
				},
				{
					UserID:   req.ReceiverID,
					Role:     ParticipantTypeMember,
					JoinedAt: timeNow,
				},
			},
			IsGroup:       false,
			CreatedAt:     timeNow,
			UpdatedAt:     timeNow,
			LastMessageAt: timeNow,
		}

		convID, err := wh.dbHandler.ConvHandler.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, err
		}
		message.ConversationID = convID
	}

	err := wh.dbHandler.MessageHandler.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	participants, err := wh.dbHandler.ConvHandler.GetConversationParticipants(ctx, message.ConversationID)
	if err != nil {
		return nil, err
	}

	participants = incrementUnreadCount(participants, userID)
	err = wh.dbHandler.ConvHandler.UpdateAddMessage(ctx, message.ConversationID, participants, timeNow)
	if err != nil {
		return nil, err
	}

	go wh.LiveSendRecipients(message.ConversationID, userID, participants, message) // Send message to live recipients via gRPC

	return &spec.AddMessageResponse{
		ConversationID: message.ConversationID,
		Message:        fmt.Sprintf("%s message added", req.MediaType),
	}, nil
}

func (wh *WassupHandler) GetMessages(ctx context.Context, req *spec.GetMessagesRequest) (*spec.GetMessagesResponse, error) {
	messages, err := wh.dbHandler.MessageHandler.GetMessages(ctx, req.ConversationID, req.Limit, req.Cursor)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func incrementUnreadCount(participants []spec.Participant, senderID string) []spec.Participant {
	for i, participant := range participants {
		if participant.UserID != senderID {
			participants[i].UnreadMessageCount++
		}
	}
	return participants
}

func (wh *WassupHandler) LiveSendRecipients(convID, userID string, participants []spec.Participant, message *spec.Message) {
	ctx := context.Background()
	for _, participant := range participants {
		if participant.UserID == userID {
			continue
		}

		recipientLive, _ := wh.redisHandler.CheckKeyPrefixExists(participant.UserID)
		log.Printf("User %s live status: %v\n", participant.UserID, recipientLive)
		if recipientLive {
			containerAddress, err := wh.redisHandler.GetContainerDetails(ctx, participant.UserID)
			if err != nil {
				fmt.Println("Error getting container details:", err)
				continue
			}

			for _, addressMap := range containerAddress {
				for key, address := range addressMap {
					fmt.Printf("Sending message to user %s at address %s\n", participant.UserID, address)
					if strings.HasSuffix(key, "name") {
						continue
					}
					err = proto.SendMessageViaGrpc(address, participant.UserID, message)
					if err != nil {
						fmt.Println("GRPC Error", err)
					}
				}
			}
		}
	}
}
