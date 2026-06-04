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
		r.Post("/users/{userID}/profile", h.SetProfileHandler)
		r.Post("/users/{userID}/profile/image", h.UploadProfileImageHandler)
		r.Post("/conversations", h.CreateConversationHandler)
		r.Post("/conversations/{convID}/message", h.AddMessageHandler)
		r.Post("/conversations/{convID}/users/{userID}/admin", h.MakeMemberAdminHandler)
		r.Post("/conversations/{convID}/users/{userID}/remove", h.RemoveMemberHandler)
		r.Post("/conversations/{convID}/users/{userID}", h.AddMemberHandler)
		r.Post("/conversations/{convID}/message/media", h.AddMediaMessageHandler)
		r.Get("/conversations/{convID}/messages", h.GetMessageHandler)
		r.Get("/conversations", h.ListConversationsHandler)
		r.Get("/ws", wsHandler.WsMessageHandler)
	})

	r.Post("/otp/send", h.SendOTPHandler)
	r.Post("/otp/verify", h.VerifyOTPHandler)

	return r
}
