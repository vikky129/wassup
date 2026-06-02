package bl

import (
	"context"
	"personal/wassup/db"
	"personal/wassup/redis"
	"personal/wassup/spec"
)

type WassupHandler struct {
	dbHandler         *db.DBHandler
	redisHandler      *redis.CacheHandler
	liveNotifyHandler LiveNotificationHandler
}

func NewWassupHandler(dbHandler *db.DBHandler, redisHandler *redis.CacheHandler, notificationHandler LiveNotificationHandler) *WassupHandler {
	return &WassupHandler{
		dbHandler:         dbHandler,
		redisHandler:      redisHandler,
		liveNotifyHandler: notificationHandler,
	}
}

type LiveNotificationHandler interface {
	//ctx context.Context, req *AddMessageRequest
	SendMessage(ctx context.Context, senderID string, participants []string, message *spec.Message) error
	UpdateParticipantStatusForConversation(ctx context.Context, convID, userID string, participants []string, status string) error
}
