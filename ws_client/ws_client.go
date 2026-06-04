package wsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"personal/wassup/redis"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn         *websocket.Conn
	messageChan  chan any
	convSvc      ConversationStatusUpdater
	closeChan    chan struct{}
	responseChan chan any
	lock         sync.Locker
	userID       string
	cacheHandler *redis.CacheHandler
}

type ConversationStatusUpdater interface {
	UpdateParticipantStatusForConversation(ctx context.Context, convID string, userID string, status string) error
}

func NewWSClient(conn *websocket.Conn, convSvc ConversationStatusUpdater, closeChan chan struct{}, userID string, cacheHandler *redis.CacheHandler) *WSClient {
	client := &WSClient{
		conn:         conn,
		messageChan:  make(chan any),
		convSvc:      convSvc,
		closeChan:    closeChan,
		userID:       userID,
		lock:         &sync.Mutex{},
		responseChan: make(chan any),
		cacheHandler: cacheHandler,
	}

	go client.readMessage(client.conn, client.closeChan)
	go sendMessage(client.conn, 5*time.Second, client.messageChan, client.closeChan, client.responseChan)

	return client
}

func (wsc *WSClient) SendMessageOnChannel(message any) any {
	wsc.lock.Lock()
	defer wsc.lock.Unlock()
	wsc.messageChan <- message
	response := <-wsc.responseChan
	return response
}

func (wsc *WSClient) readMessage(conn *websocket.Conn, closeChan chan struct{}) {
	for {
		messType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("failed reading message from websocket:", err)
			closeChan <- struct{}{}
			return
		}

		if messType == websocket.TextMessage {
			var messageRead MessageRead
			err = json.Unmarshal(message, &messageRead)

			if err != nil {
				fmt.Println("failed unmarshaling message read from websocket:", err)
				continue
			}

			if messageRead.ConversationID == "" || messageRead.UserID == "" {
				fmt.Println("invalid message read from websocket: missing conversation ID or user ID")
				continue
			}

			if messageRead.MessageType == MessageDeliveryStatus || messageRead.MessageType == MessageReadStatus {
				err = wsc.convSvc.UpdateParticipantStatusForConversation(context.Background(), messageRead.ConversationID, messageRead.UserID, string(messageRead.MessageType))
				if err != nil {
					fmt.Println("failed updating participant status for conversation:", err)
					continue
				}
				fmt.Printf("Successfully updated participant status to %s for user %s in conversation %s\n", messageRead.MessageType, messageRead.UserID, messageRead.ConversationID)
				continue
			}
		}
	}
}

func sendMessage(conn *websocket.Conn, interval time.Duration, messageChan chan any, closeChan chan struct{}, responseChan chan any) {
	tick := time.NewTicker(interval)

	for {
		select {
		case <-tick.C:
			err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(30*time.Second))
			if err != nil {
				fmt.Println("failed sending ping message to websocket:", err)
				closeChan <- struct{}{}
				return
			}
		case message := <-messageChan:
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				fmt.Println("failed marshaling message to JSON:", err)
				break
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err = conn.WriteMessage(websocket.TextMessage, jsonMessage)
			if err != nil {
				var netErr net.Error
				if ok := errors.As(err, &netErr); ok && netErr.Timeout() {
					fmt.Println("write message to websocket timed out:", err)
					closeChan <- struct{}{}
					return
				}
				responseChan <- fmt.Errorf("Error sending message: %v", err)
				fmt.Println("failed writing message to websocket:", err)
				break
			}
			responseChan <- "Message sent successfully"
		}
	}
}
