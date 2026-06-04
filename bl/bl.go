package bl

import (
	"context"
	"personal/wassup/spec"
)

type DBHandler interface {
	CreateUser(ctx context.Context, user *spec.User) (string, error)
	GetUserByID(ctx context.Context, userID string) (*spec.User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*spec.User, error)
	UpdateUser(ctx context.Context, user *spec.User) error
	DeleteUser(ctx context.Context, userID string) error

	GetConversationParticipants(ctx context.Context, convID string) ([]spec.Participant, error)
	CreateConversation(ctx context.Context, conv *spec.Conversation) (string, error)
	UpdateConversationParticipants(ctx context.Context, convID string, participants []spec.Participant) error
	UpdateAddMessage(ctx context.Context, convID string, participants []spec.Participant, updatedAt int64) error
	ListConversations(ctx context.Context, userID string) ([]spec.Conversation, error)
	GetUnReadConversations(ctx context.Context, userID string) ([]string, error)
	UpdateConversationParticipantStatus(ctx context.Context, convID, userID, status string) error

	AddMessage(ctx context.Context, msg *spec.Message) error
	GetMessages(ctx context.Context, convID string, limit int, cursor string) (*spec.GetMessagesResponse, error)
	UpdateStatusDelivered(ctx context.Context, convIDs []string, userID string) error
	UpdateStatusRead(ctx context.Context, convID string, userID string) error
}

type RedisHandler interface {
	Set(key string, otp any) error
	Get(key string) (string, error)
}

type LiveNotificationHandler interface {
	SendMessage(ctx context.Context, senderID string, participants []string, message *spec.Message) error
	UpdateParticipantStatusForConversation(ctx context.Context, convID, userID string, participants []string, status string) error
}

type TokenService interface {
	GenerateToken(userID string) (string, error)
}

type AuthService interface {
	SendOTP(ctx context.Context, otpRequest *spec.SendOtpRequest) (*spec.SendOtpResponse, error)
	VerifyOTP(ctx context.Context, otpRequest *spec.VerifyOtpRequest) (*spec.VerifyOtpResponse, error)
}

type UserService interface {
	SetProfile(ctx context.Context, setProfileRequest *spec.SetProfileRequest) (*spec.SetProfileResponse, error)
	UploadProfileImage(ctx context.Context, userID, filePath string) error
}

type ConversationService interface {
	CreateGroupConversation(ctx context.Context, req *spec.CreateConversationRequest) (string, error)
	MakeMemberAdmin(ctx context.Context, convID string, userID string) error
	RemoveMember(ctx context.Context, convID string, userID string) error
	AddMember(ctx context.Context, convID string, userID string) error
	ListConversations(ctx context.Context) ([]spec.Conversation, error)
	UpdateParticipantStatusForConversation(ctx context.Context, convID string, userID string, status string) error
}

type MessageService interface {
	AddMessage(ctx context.Context, req *spec.AddMessageRequest) (*spec.AddMessageResponse, error)
	AddMediaMessage(ctx context.Context, req *spec.AddMediaMessageRequest, filePath string) (*spec.AddMessageResponse, error)
	GetMessages(ctx context.Context, req *spec.GetMessagesRequest) (*spec.GetMessagesResponse, error)
}

type Services struct {
	Auth         AuthService
	User         UserService
	Conversation ConversationService
	Message      MessageService
}

type authService struct {
	dbHandler    DBHandler
	redisHandler RedisHandler
	tokenService TokenService
}

type userService struct {
	dbHandler DBHandler
}

type conversationService struct {
	dbHandler         DBHandler
	liveNotifyHandler LiveNotificationHandler
}

type messageService struct {
	dbHandler         DBHandler
	liveNotifyHandler LiveNotificationHandler
}

func NewServices(dbHandler DBHandler, redisHandler RedisHandler, notificationHandler LiveNotificationHandler, tokenService TokenService) *Services {
	return &Services{
		Auth: &authService{
			dbHandler:    dbHandler,
			redisHandler: redisHandler,
			tokenService: tokenService,
		},
		User: &userService{
			dbHandler: dbHandler,
		},
		Conversation: &conversationService{
			dbHandler:         dbHandler,
			liveNotifyHandler: notificationHandler,
		},
		Message: &messageService{
			dbHandler:         dbHandler,
			liveNotifyHandler: notificationHandler,
		},
	}
}
