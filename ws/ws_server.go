package ws

import (
	"fmt"
	"net/http"
	"personal/wassup/db"
	"personal/wassup/redis"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	dbHandler  *db.DBHandler
	localcache map[string]map[string]*WSClient
	redisCache *redis.CacheHandler
}

func NewWebSocketHandler(db *db.DBHandler, redisHandler *redis.CacheHandler) *WebSocketHandler {
	return &WebSocketHandler{
		dbHandler:  db,
		localcache: make(map[string]map[string]*WSClient),
		redisCache: redisHandler,
	}
}

func (h *WebSocketHandler) GetWSClients(userID string) ([]*WSClient, error) {
	fmt.Println("local cache:", h.localcache)
	if clientMap, ok := h.localcache[userID]; ok {
		clients := make([]*WSClient, 0, len(clientMap))
		for _, client := range clientMap {
			clients = append(clients, client)
		}
		return clients, nil
	} else {
		return nil, fmt.Errorf("userID not found in local cache")
	}
}

func (h *WebSocketHandler) WsMessageHandler(w http.ResponseWriter, req *http.Request) {
	deviceID := req.URL.Query().Get("deviceID")
	userID, ok := req.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "user_id not found in context", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}

	closeConnChan := make(chan struct{})

	wsClient := NewWSClient(conn, h.dbHandler, closeConnChan)

	go func() {
		<-closeConnChan
		delete(h.localcache[userID], deviceID)
		if len(h.localcache[userID]) == 0 {
			delete(h.localcache, userID)
		}

		err = h.redisCache.DeleteKey(fmt.Sprintf("%s:%s", userID, deviceID))
		if err != nil {
			fmt.Println("Error deleting key from redis:", err)
		}
	}()

	if _, exists := h.localcache[userID]; !exists {
		h.localcache[userID] = make(map[string]*WSClient)
	}
	h.localcache[userID][deviceID] = wsClient
	fmt.Println("local cache after adding client:", h.localcache)
	//h.redisCache.Set(fmt.Sprintf("%s:%s", userID, deviceID), containerID)
	h.redisCache.StoreContainerDetails(req.Context(), userID, deviceID)

	unreadConversations, err := h.dbHandler.ConvHandler.GetUnReadConversations(req.Context(), userID)
	if err != nil {
		fmt.Println("Error getting unread conversations:", err)
	}

	if len(unreadConversations) > 0 {
		err = h.dbHandler.MessageHandler.UpdateStatusDelivered(req.Context(), unreadConversations, userID)
		if err != nil {
			fmt.Println("Error updating message status to delivered:", err)
		}
	}
}
