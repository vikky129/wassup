package bl

import (
	"context"
	"fmt"
	"personal/wassup/spec"
	"time"
)

const (
	ParticipantTypeAdmin  = "Admin"
	ParticipantTypeMember = "Member"
)

func (svc *conversationService) CreateGroupConversation(ctx context.Context, req *spec.CreateConversationRequest) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user_id not found in context")
	}

	adminParticipant := spec.Participant{
		UserID: userID,
		Role:   ParticipantTypeAdmin,
	}

	req.Participants = append(req.Participants, adminParticipant)

	for i := range req.Participants {
		req.Participants[i].JoinedAt = time.Now().Unix()
	}

	var conversation = &spec.Conversation{
		Participants: req.Participants,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	if req.IsGroup {
		conversation.IsGroup = true
		conversation.Name = req.Name
		conversation.Description = req.Description
	}

	convID, err := svc.dbHandler.CreateConversation(ctx, conversation)
	if err != nil {
		return "", err
	}

	return convID, nil
}

func (svc *conversationService) MakeMemberAdmin(ctx context.Context, convID string, userID string) error {
	participants, err := svc.dbHandler.GetConversationParticipants(ctx, convID)
	if err != nil {
		return err
	}

	var updated bool
	for i := range participants {
		if participants[i].UserID == userID {
			participants[i].Role = ParticipantTypeAdmin
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("user not found in conversation")
	}

	return svc.dbHandler.UpdateConversationParticipants(ctx, convID, participants)
}

func (svc *conversationService) RemoveMember(ctx context.Context, convID string, userID string) error {

	currUserID, ok := ctx.Value("user_id").(string)
	if !ok || currUserID == "" {
		return fmt.Errorf("user_id not found in context")
	}
	participants, err := svc.dbHandler.GetConversationParticipants(ctx, convID)
	if err != nil {
		return err
	}

	var isCurrentUserAdmin bool
	for i := range participants {
		if participants[i].UserID == currUserID && participants[i].Role == ParticipantTypeAdmin {
			isCurrentUserAdmin = true
			break
		}
	}

	if !isCurrentUserAdmin {
		return fmt.Errorf("only admins can remove members from the conversation")
	}

	var updatedParticipants []spec.Participant
	for i := range participants {
		if participants[i].UserID != userID {
			updatedParticipants = append(updatedParticipants, participants[i])
		}
	}

	return svc.dbHandler.UpdateConversationParticipants(ctx, convID, updatedParticipants)
}

func (svc *conversationService) AddMember(ctx context.Context, convID string, userID string) error {
	participants, err := svc.dbHandler.GetConversationParticipants(ctx, convID)
	if err != nil {
		return err
	}

	for i := range participants {
		if participants[i].UserID == userID {
			return fmt.Errorf("user is already a participant in the conversation")
		}
	}

	newParticipant := spec.Participant{
		UserID:   userID,
		Role:     ParticipantTypeMember,
		JoinedAt: time.Now().Unix(),
	}

	participants = append(participants, newParticipant)

	return svc.dbHandler.UpdateConversationParticipants(ctx, convID, participants)
}

func (svc *conversationService) ListConversations(ctx context.Context) ([]spec.Conversation, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	return svc.dbHandler.ListConversations(ctx, userID)
}

func (svc *conversationService) UpdateParticipantStatusForConversation(ctx context.Context, convID string, userID string, status string) error {
	err := svc.dbHandler.UpdateConversationParticipantStatus(context.Background(), convID, userID, status)
	if err != nil {
		fmt.Println("failed updating conversation participant status:", err)
		return err
	}

	participants, err := svc.dbHandler.GetConversationParticipants(ctx, convID)
	if err != nil {
		return err
	}

	var userIDs []string
	for _, participant := range participants {
		userIDs = append(userIDs, participant.UserID)
	}

	go svc.liveNotifyHandler.UpdateParticipantStatusForConversation(ctx, convID, userID, userIDs, status)

	return nil
}
