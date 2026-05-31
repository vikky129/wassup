package transport

import (
	"personal/wassup/middleware"
	"personal/wassup/ws"

	"github.com/go-chi/chi/v5"
)

func Router(h *HandlerStruct, wsHandler *ws.WebSocketHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Post("/set_profile", h.SetProfileHandler)
		r.Post("/upload_profile_image", h.UploadProfileImageHandler)
		r.Post("/create_conversation", h.CreateConversationHandler)
		r.Post("/add_message", h.AddMessageHandler)
		r.Post("/make_member_admin", h.MakeMemberAdminHandler)
		r.Post("/remove_member", h.RemoveMemberHandler)
		r.Post("/add_member", h.AddMemberHandler)
		r.Post("/add_media_message", h.AddMediaMessageHandler)
		r.Get("/conversations/{convID}/messages", h.GetMessageHandler)
		r.Get("/list_conversations", h.ListConversationsHandler)
		r.Get("/ws", wsHandler.WsMessageHandler)
	})

	r.Post("/send_otp", h.SendOTPHandler)
	r.Post("/verify_otp", h.VerifyOTPHandler)

	return r
}
