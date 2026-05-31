package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"personal/wassup/db"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn         *websocket.Conn
	messageChan  chan any
	dbHandler    *db.DBHandler
	closeChan    chan struct{}
	responseChan chan any
	lock         sync.Locker
}

func NewWSClient(conn *websocket.Conn, dbHandler *db.DBHandler, closeChan chan struct{}) *WSClient {
	client := &WSClient{
		conn:         conn,
		messageChan:  make(chan any),
		dbHandler:    dbHandler,
		closeChan:    closeChan,
		lock:         &sync.Mutex{},
		responseChan: make(chan any),
	}

	go readMessage(client.conn, client.closeChan)
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

func readMessage(conn *websocket.Conn, closeChan chan struct{}) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("failed reading message from websocket:", err)
			closeChan <- struct{}{}
			return
		}
		println("Received:", string(message))
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
