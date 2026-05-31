package ws

import "personal/wassup/spec"

type MessageType string

const (
	AddMessage            MessageType = "AddMessage"
	MessageDeliveryStatus MessageType = "MessageDeliveryStatus"
)

type MessageDelivery struct {
	MessageID      string `json:"messageId"`
	ConversationID string `json:"conversationId"`
	Status         string `json:"status"`
}

type WSMessage struct {
	MessageType           MessageType     `json:"messageType"`
	AddMessagePayload     spec.Message    `json:"addMessagePayload"`
	MessageDeliveryStatus MessageDelivery `json:"messageDeliveryStatus"`
}
