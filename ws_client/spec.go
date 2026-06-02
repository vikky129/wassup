package wsclient

type MessageType string

const (
	AddMessage            MessageType = "AddMessage"
	MessageDeliveryStatus MessageType = "MessageDelivered"
	MessageReadStatus     MessageType = "MessageRead"
)

type MessageRead struct {
	ConversationID string      `json:"conversationId"`
	MessageType    MessageType `json:"messageType"`
	UserID         string      `json:"userId"`
}
