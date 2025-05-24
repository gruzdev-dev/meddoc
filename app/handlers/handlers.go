package handlers

import (
	"github.com/gorilla/mux"
	"github.com/gruzdev-dev/meddoc/app/services/user"
)

type Handlers struct {
	userHandler *UserHandler
}

func NewHandlers(userService *user.UserService) *Handlers {
	return &Handlers{
		userHandler: NewUserHandler(userService),
	}
}

func (h *Handlers) RegisterRoutes(router *mux.Router) {
	h.userHandler.RegisterRoutes(router)
}
