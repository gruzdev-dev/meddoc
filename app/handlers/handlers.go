package handlers

import (
	"github.com/gorilla/mux"
	"github.com/gruzdev-dev/meddoc/app/services/document"
	"github.com/gruzdev-dev/meddoc/app/services/user"
)

type Handlers struct {
	userHandler     *UserHandler
	documentHandler *DocumentHandler
}

func NewHandlers(userService *user.UserService, documentService *document.Service) *Handlers {
	return &Handlers{
		userHandler:     NewUserHandler(userService),
		documentHandler: NewDocumentHandler(documentService, userService),
	}
}

func (h *Handlers) RegisterRoutes(router *mux.Router) {
	h.userHandler.RegisterRoutes(router)
	h.documentHandler.RegisterRoutes(router)
}
