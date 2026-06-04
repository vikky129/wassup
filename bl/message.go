package bl

import (
	"context"
	"fmt"
	"personal/wassup/spec"
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

func (svc *messageService) AddMessage(ctx context.Context, req *spec.AddMessageRequest) (*spec.AddMessageResponse, error) {
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

		convID, err := svc.dbHandler.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, err
		}
		message.ConversationID = convID
		participants = conversation.Participants
	}

	err := svc.dbHandler.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	if len(participants) == 0 {
		participants, err = svc.dbHandler.GetConversationParticipants(ctx, req.ConversationID)
		if err != nil {
			return nil, err
		}
	}

	participants = incrementUnreadCount(participants, userID)
	err = svc.dbHandler.UpdateAddMessage(ctx, req.ConversationID, participants, timeNow)
	if err != nil {
		return nil, err
	}

	var participantList []string
	for _, participant := range participants {
		if participant.UserID != userID {
			participantList = append(participantList, participant.UserID)
		}
	}

	go svc.liveNotifyHandler.SendMessage(context.Background(), userID, participantList, message) // Send message to live recipients via gRPC

	//go svc.LiveSendRecipients(message.ConversationID, userID, participants, message) // Send message to live recipients via gRPC

	return &spec.AddMessageResponse{
		ConversationID: message.ConversationID,
		Message:        message.Text,
	}, nil
}

func (svc *messageService) AddMediaMessage(ctx context.Context, req *spec.AddMediaMessageRequest, filePath string) (*spec.AddMessageResponse, error) {
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

		convID, err := svc.dbHandler.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, err
		}
		message.ConversationID = convID
	}

	err := svc.dbHandler.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	participants, err := svc.dbHandler.GetConversationParticipants(ctx, message.ConversationID)
	if err != nil {
		return nil, err
	}

	participants = incrementUnreadCount(participants, userID)
	err = svc.dbHandler.UpdateAddMessage(ctx, message.ConversationID, participants, timeNow)
	if err != nil {
		return nil, err
	}

	//go svc.LiveSendRecipients(message.ConversationID, userID, participants, message) // Send message to live recipients via gRPC

	return &spec.AddMessageResponse{
		ConversationID: message.ConversationID,
		Message:        fmt.Sprintf("%s message added", req.MediaType),
	}, nil
}

func (svc *messageService) GetMessages(ctx context.Context, req *spec.GetMessagesRequest) (*spec.GetMessagesResponse, error) {
	messages, err := svc.dbHandler.GetMessages(ctx, req.ConversationID, req.Limit, req.Cursor)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func incrementUnreadCount(participants []spec.Participant, senderID string) []spec.Participant {
	for i, participant := range participants {
		if participant.UserID != senderID {
			participants[i].UnreadMessageCount++
			participants[i].MessageStatus = spec.StatusSent
		}
	}
	return participants
}
