package bl

import (
	"personal/wassup/db"
	"personal/wassup/redis"
)

type WassupHandler struct {
	dbHandler    *db.DBHandler
	redisHandler *redis.CacheHandler
}

func NewWassupHandler(dbHandler *db.DBHandler, redisHandler *redis.CacheHandler) *WassupHandler {
	return &WassupHandler{
		dbHandler:    dbHandler,
		redisHandler: redisHandler,
	}
}
