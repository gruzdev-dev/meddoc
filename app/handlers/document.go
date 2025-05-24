package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gruzdev-dev/meddoc/app/models"
	"github.com/gruzdev-dev/meddoc/app/server/context"
	"github.com/gruzdev-dev/meddoc/app/server/middleware"
	"github.com/gruzdev-dev/meddoc/app/services/document"
	"github.com/gruzdev-dev/meddoc/app/services/user"
)

type DocumentHandler struct {
	documentService *document.Service
	userService     *user.UserService
}

func NewDocumentHandler(documentService *document.Service, userService *user.UserService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		userService:     userService,
	}
}

func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var doc models.DocumentCreation
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := context.GetUserID(r)
	createdDoc, err := h.documentService.CreateDocument(r.Context(), doc, userID)
	if err != nil {
		http.Error(w, "failed to create document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdDoc)
}

func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := context.GetUserID(r)

	doc, err := h.documentService.GetDocument(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, document.ErrAccessDenied) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentHandler) GetUserDocuments(w http.ResponseWriter, r *http.Request) {
	userID := context.GetUserID(r)
	docs, err := h.documentService.GetUserDocuments(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := context.GetUserID(r)

	if err := h.documentService.DeleteDocument(r.Context(), id, userID); err != nil {
		if errors.Is(err, document.ErrAccessDenied) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		if errors.Is(err, document.ErrDocumentNotFound) {
			http.Error(w, "document not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete document", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := context.GetUserID(r)

	var update models.DocumentUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedDoc, err := h.documentService.UpdateDocument(r.Context(), id, update, userID)
	if err != nil {
		if errors.Is(err, document.ErrAccessDenied) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		if errors.Is(err, document.ErrDocumentNotFound) {
			http.Error(w, "document not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedDoc)
}

func (h *DocumentHandler) RegisterRoutes(router *mux.Router) {
	docs := router.PathPrefix("/documents").Subrouter()
	docs.Use(middleware.Auth(h.userService))

	docs.HandleFunc("", h.CreateDocument).Methods(http.MethodPost)
	docs.HandleFunc("", h.GetUserDocuments).Methods(http.MethodGet)
	docs.HandleFunc("/{id}", h.GetDocument).Methods(http.MethodGet)
	docs.HandleFunc("/{id}", h.UpdateDocument).Methods(http.MethodPatch)
	docs.HandleFunc("/{id}", h.DeleteDocument).Methods(http.MethodDelete)
}
