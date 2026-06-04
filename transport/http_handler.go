package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"personal/wassup/bl"
	"personal/wassup/media"
	"personal/wassup/spec"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type HandlerStruct struct {
	authService         bl.AuthService
	userService         bl.UserService
	conversationService bl.ConversationService
	messageService      bl.MessageService
	mediaService        media.MediaRepository
}

func NewHandler(services *bl.Services, mediaService media.MediaRepository) *HandlerStruct {
	return &HandlerStruct{
		authService:         services.Auth,
		userService:         services.User,
		conversationService: services.Conversation,
		messageService:      services.Message,
		mediaService:        mediaService,
	}
}

func (h *HandlerStruct) Close() error {
	// Clean up any resources if needed
	return nil
}

func (h *HandlerStruct) SendOTPHandler(res http.ResponseWriter, req *http.Request) {
	var reqObj spec.SendOtpRequest
	if err := json.NewDecoder(req.Body).Decode(&reqObj); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	respObj, err := h.authService.SendOTP(req.Context(), &reqObj)
	if err != nil {
		http.Error(res, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, err)
}

func (h *HandlerStruct) VerifyOTPHandler(res http.ResponseWriter, req *http.Request) {
	var verifyReq spec.VerifyOtpRequest
	if err := json.NewDecoder(req.Body).Decode(&verifyReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	respObj, err := h.authService.VerifyOTP(req.Context(), &verifyReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to verify OTP: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, err)
	// Handle verifying OTP
}

func (h *HandlerStruct) SetProfileHandler(res http.ResponseWriter, req *http.Request) {
	var profileReq spec.SetProfileRequest
	if err := json.NewDecoder(req.Body).Decode(&profileReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	respObj, err := h.userService.SetProfile(req.Context(), &profileReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to set profile: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, err)
}

func (h *HandlerStruct) UploadProfileImageHandler(res http.ResponseWriter, req *http.Request) {

	image, header, err := req.FormFile("image")
	if err != nil {
		http.Error(res, fmt.Sprintf("Error retrieving file: %v", err), http.StatusBadRequest)
		return
	}

	defer image.Close()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to create uploads directory: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := "./uploads/" + header.Filename
	fileH, err := os.Create(filePath)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	defer fileH.Close()

	_, err = io.Copy(fileH, image)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	userID, ok := req.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(res, "user_id not found in context", http.StatusUnauthorized)
		return
	}

	err = h.userService.UploadProfileImage(req.Context(), userID, filePath)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to upload profile image: %v", err), http.StatusInternalServerError)
		return
	}

	handleResponse(res, "image uploaded successfully", nil)
}

func (h *HandlerStruct) CreateConversationHandler(res http.ResponseWriter, req *http.Request) {
	var createConvReq spec.CreateConversationRequest
	if err := json.NewDecoder(req.Body).Decode(&createConvReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received CreateConversationRequest: %+v\n", createConvReq)

	respObj, err := h.conversationService.CreateGroupConversation(req.Context(), &createConvReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to create conversation: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, err)
}

func (h *HandlerStruct) AddMessageHandler(res http.ResponseWriter, req *http.Request) {
	var addMsgReq spec.AddMessageRequest
	if err := json.NewDecoder(req.Body).Decode(&addMsgReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.messageService.AddMessage(req.Context(), &addMsgReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to add message: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, resp, err)
}

func (h *HandlerStruct) MakeMemberAdminHandler(res http.ResponseWriter, req *http.Request) {
	var updateReq spec.UpdateMemberRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.conversationService.MakeMemberAdmin(req.Context(), updateReq.ConversationID, updateReq.UserID)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to make member admin: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, "member promoted to admin successfully", nil)
}

func (h *HandlerStruct) RemoveMemberHandler(res http.ResponseWriter, req *http.Request) {
	// Handle removing member from conversation
	var updateReq spec.UpdateMemberRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.conversationService.RemoveMember(req.Context(), updateReq.ConversationID, updateReq.UserID)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to remove member: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, "member removed successfully", nil)
}

func (h *HandlerStruct) AddMemberHandler(res http.ResponseWriter, req *http.Request) {
	// Handle adding member to conversation
	var updateReq spec.UpdateMemberRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.conversationService.AddMember(req.Context(), updateReq.ConversationID, updateReq.UserID)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to add member: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, "member added successfully", nil)
}

func (h *HandlerStruct) ListConversationsHandler(res http.ResponseWriter, req *http.Request) {
	respObj, err := h.conversationService.ListConversations(req.Context())
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to list conversations: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, err)
}

func (h *HandlerStruct) AddMediaMessageHandler(res http.ResponseWriter, req *http.Request) {
	isNewConversation := req.FormValue("is_new_conversation") == "true"
	receiverID := req.FormValue("receiver_id") // needed if it's a new conversation

	convID := req.FormValue("conversation_id")

	if convID == "" && !isNewConversation {
		http.Error(res, "conversation_id is required for existing conversations", http.StatusBadRequest)
		return
	}

	if isNewConversation && receiverID == "" {
		http.Error(res, "receiver_id is required for new conversations", http.StatusBadRequest)
		return
	}

	if convID != "" && isNewConversation {
		http.Error(res, "conversation_id should not be provided for new conversations", http.StatusBadRequest)
		return
	}

	file, header, err := req.FormFile("media")
	if err != nil {
		http.Error(res, fmt.Sprintf("Error retrieving file: %v", err), http.StatusBadRequest)
		return
	}

	defer file.Close()

	if header.Size > 10*1024*1024 { // 10MB limit
		http.Error(res, "File size exceeds 10MB limit", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads/message", os.ModePerm)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error creating upload directory: %v", err), http.StatusInternalServerError)
		return
	}

	messageType := getMediaType(header)

	var fileBytes []byte
	_, err = file.Read(fileBytes)
	if err != nil {
		http.Error(res, "error reading file bytes", http.StatusBadRequest)
		return
	}
	path, err := h.mediaService.UploadMedia(req.Context(), fileBytes, header.Filename, messageType)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to upload with error: %v", err), http.StatusInternalServerError)
	}

	addMsgReq := &spec.AddMediaMessageRequest{
		IsNewConversation: isNewConversation,
		ConversationID:    convID,
		ReceiverID:        receiverID,
		MediaType:         messageType,
	}

	resp, err := h.messageService.AddMediaMessage(req.Context(), addMsgReq, path)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to add media message: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, resp, nil)
}

func (h *HandlerStruct) GetMessageHandler(res http.ResponseWriter, req *http.Request) {
	var getMsgReq spec.GetMessagesRequest
	var convID = chi.URLParam(req, "convID")
	limitStr := req.URL.Query().Get("limit")
	cursor := req.URL.Query().Get("cursor")

	if convID == "" {
		http.Error(res, "conversation_id is required", http.StatusBadRequest)
		return
	}

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(res, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		getMsgReq.Limit = limit
	}

	getMsgReq.ConversationID = convID
	getMsgReq.Cursor = cursor

	respObj, err := h.messageService.GetMessages(req.Context(), &getMsgReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Failed to get messages: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	handleResponse(res, respObj, nil)
}

func getMediaType(header *multipart.FileHeader) string {
	cType := header.Header.Get("Content-Type")
	if strings.HasPrefix(cType, "image/") {
		return bl.MessageTypeImage
	}

	if strings.HasPrefix(cType, "video/") {
		return bl.MessageTypeVideo
	}

	if strings.HasPrefix(cType, "audio/") {
		return bl.MessageTypeAudio
	}

	return bl.MessageTypeFile
}
