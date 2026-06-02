package spec

type User struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string `bson:"name" json:"name"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	CreatedAt   int64  `bson:"created_at" json:"created_at"`
	UpdatedAt   int64  `bson:"updated_at" json:"updated_at"`
	LastOnline  int64  `bson:"last_online" json:"last_online"`
	IsVerified  bool   `bson:"is_verified" json:"is_verified"`
	PhotoURL    string `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
	Bio         string `bson:"bio,omitempty" json:"bio,omitempty"`
}

type Conversation struct {
	ID            string        `bson:"_id" json:"id"`
	Participants  []Participant `bson:"participants" json:"participants"`
	CreatedAt     int64         `bson:"created_at" json:"created_at"`
	UpdatedAt     int64         `bson:"updated_at" json:"updated_at"`
	Name          string        `bson:"name" json:"name"`
	PhotoURL      string        `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
	Description   string        `bson:"description,omitempty" json:"description,omitempty"`
	IsGroup       bool          `bson:"is_group" json:"is_group"`
	LastMessageAt int64         `bson:"last_message_at" json:"last_message_at"`
}

type Participant struct {
	UserID             string                `bson:"user_id" json:"user_id"`
	JoinedAt           int64                 `bson:"joined_at" json:"joined_at"`
	UnreadMessageCount int                   `bson:"unread_message_count" json:"unread_message_count"`
	Role               string                `bson:"role" json:"role"` // "admin" or "member"
	MessageStatus      MessageDeliveryStatus `bson:"message_status" json:"message_status"`
}

// #################       User-facing API Request/Response Structs       #################
type SendOtpRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type SendOtpResponse struct {
	OTP int64 `json:"otp"`
}

type VerifyOtpRequest struct {
	PhoneNumber string `json:"phone_number"`
	Otp         string `json:"otp"`
}

type VerifyOtpResponse struct {
	Success bool   `json:"success"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
}

type SetProfileRequest struct {
	Name string `json:"name"`
	Bio  string `json:"bio,omitempty"`
}

type SetProfileResponse struct {
	Result string `json:"result"`
}

// #################       Conversation Management API Request/Response Structs       #################
type CreateConversationRequest struct {
	Participants []Participant `json:"participants"`
	IsGroup      bool          `json:"is_group"`
	Name         string        `json:"name,omitempty"`
	Description  string        `json:"description,omitempty"`
}

type CreateConversationResponse struct {
	ConversationID string `json:"conversation_id"`
}

type UpdateMemberRequest struct {
	UserID         string `json:"user_id"`
	ConversationID string `json:"conversation_id"`
}

// #################       Message Management API Request/Response Structs       #################

type MessageDeliveryStatus string

const (
	StatusSent      MessageDeliveryStatus = "sent"
	StatusDelivered MessageDeliveryStatus = "delivered"
	StatusRead      MessageDeliveryStatus = "read"
)

type Message struct {
	ID             string `bson:"_id" json:"id"`
	ConversationID string `bson:"conversation_id" json:"conversation_id"`
	SenderID       string `bson:"sender_id" json:"sender_id"`
	Text           string `bson:"text" json:"text"`
	Type           string `bson:"type" json:"type"` // "audio", "image", "video", "document"
	MediaURL       string `bson:"media_url,omitempty" json:"media_url,omitempty"`
	CreatedAt      int64  `bson:"created_at" json:"created_at"`
	UpdatedAt      int64  `bson:"updated_at" json:"updated_at"`
}

type AddMessageRequest struct {
	IsNewConversation bool   `json:"is_new_conversation"`
	Text              string `json:"text,omitempty"`
	ConversationID    string `json:"conversation_id,omitempty"`
	ReceiverID        string `json:"receiver_id"`
}

type AddMessageResponse struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type AddMediaMessageRequest struct {
	IsNewConversation bool   `json:"is_new_conversation"` // If true, ConversationID can be empty and ReceiverID must be provided to create a new conversation
	ConversationID    string `json:"conversation_id,omitempty"`
	ReceiverID        string `json:"receiver_id"` // Needed to create new conversation if IsNewConversation is true
	MediaType         string `json:"media_type"`  // "image", "video", "file"
}

type GetMessagesRequest struct {
	ConversationID string `json:"conversation_id"`
	Limit          int    `json:"limit"`
	Cursor         string `json:"cursor,omitempty"`
}

type GetMessagesResponse struct {
	Messages   []Message `json:"messages"`
	NextCursor string    `json:"next_cursor,omitempty"`
}
