package handlers

import (
	"github.com/gorilla/mux"
	"github.com/gruzdev-dev/meddoc/app/services/document"
	"github.com/gruzdev-dev/meddoc/app/services/file"
	"github.com/gruzdev-dev/meddoc/app/services/user"
)

type Handlers struct {
	userHandler     *UserHandler
	documentHandler *DocumentHandler
	fileHandler     *FileHandler
}

func NewHandlers(userService *user.UserService, documentService *document.Service, fileService *file.Service) *Handlers {
	return &Handlers{
		userHandler:     NewUserHandler(userService),
		documentHandler: NewDocumentHandler(documentService, userService),
		fileHandler:     NewFileHandler(fileService, userService),
	}
}

func (h *Handlers) RegisterRoutes(router *mux.Router) {
	h.userHandler.RegisterRoutes(router)
	h.documentHandler.RegisterRoutes(router)
	h.fileHandler.RegisterRoutes(router)
}
