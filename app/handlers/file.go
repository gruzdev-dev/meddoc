package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
	"github.com/gruzdev-dev/meddoc/app/server/context"
	"github.com/gruzdev-dev/meddoc/app/server/middleware"
	"github.com/gruzdev-dev/meddoc/app/services/file"
	"github.com/gruzdev-dev/meddoc/app/services/user"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
)

var allowedMimeTypes = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/jpg":       {},
	"image/png":       {},
}

var allowedExts = map[string]struct{}{
	".pdf":  {},
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

type FileHandler struct {
	fileService *file.Service
	userService *user.UserService
}

func NewFileHandler(fileService *file.Service, userService *user.UserService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		userService: userService,
	}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	const maxSize = 100 << 20 // 100MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	fileReader, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "error retrieving file", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			logger.Error("failed to close file", err)
		}
	}()

	buffer := make([]byte, 512)
	n, err := fileReader.Read(buffer)
	if err != nil && err != io.EOF {
		http.Error(w, "error reading file", http.StatusBadRequest)
		return
	}
	buffer = buffer[:n]

	mimeType := header.Header.Get("Content-Type")
	detectedType := http.DetectContentType(buffer)
	ext := strings.ToLower(filepath.Ext(header.Filename))

	if !isValidMimeType(mimeType) || !isValidExtension(ext) || !isValidMimeType(detectedType) {
		http.Error(w, "invalid file type", http.StatusBadRequest)
		return
	}

	multiReader := io.MultiReader(bytes.NewReader(buffer), fileReader)

	userID := context.GetUserID(r)
	logger.Info("uploading file", map[string]any{
		"user_id": userID,
	})

	metadata := models.FileMetadata{
		Size: header.Size,
	}

	uploadedFile, err := h.fileService.UploadFile(r.Context(), multiReader, metadata, userID)
	if err != nil {
		logger.Error("failed to upload file", err)
		http.Error(w, "failed to upload file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(uploadedFile); err != nil {
		logger.Error("failed to encode response", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func isValidMimeType(mimeType string) bool {
	_, exists := allowedMimeTypes[mimeType]
	return exists
}

func isValidExtension(ext string) bool {
	_, exists := allowedExts[ext]
	return exists
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := context.GetUserID(r)

	reader, err := h.fileService.DownloadFile(r.Context(), id, userID)
	if err != nil {
		logger.Error("failed to download file", err)
		switch err {
		case errors.ErrAccessDenied:
			http.Error(w, "access denied", http.StatusForbidden)
		case errors.ErrNotFound:
			http.Error(w, "file not found", http.StatusNotFound)
		default:
			http.Error(w, "failed to download file", http.StatusInternalServerError)
		}
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("failed to close reader", err)
		}
	}()

	if _, err := io.Copy(w, reader); err != nil {
		logger.Error("failed to send file", err)
		http.Error(w, "failed to send file", http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) RegisterRoutes(router *mux.Router) {
	files := router.PathPrefix("/files").Subrouter()
	files.Use(middleware.Auth(h.userService))

	files.HandleFunc("/upload", h.UploadFile).Methods(http.MethodPost)
	files.HandleFunc("/{id}", h.DownloadFile).Methods(http.MethodGet)
}
