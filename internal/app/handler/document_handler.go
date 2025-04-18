package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appservice "github.com/gruzdev-dev/meddoc/internal/app/service"
	"github.com/gruzdev-dev/meddoc/internal/domain"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentHandler struct {
	service *appservice.DocumentService
}

func NewDocumentHandler(service *appservice.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		service: service,
	}
}

func (h *DocumentHandler) RegisterRoutes(router *gin.RouterGroup) {
	documents := router.Group("/documents")
	{
		documents.POST("", h.CreateDocument)
		documents.GET("", h.GetAllDocuments)
		documents.GET("/:id", h.GetDocument)
		documents.PUT("/:id", h.UpdateDocument)
		documents.DELETE("/:id", h.DeleteDocument)
	}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var doc domain.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		logger.Error("invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.CreateDocument(c.Request.Context(), &doc); err != nil {
		switch err {
		case domain.ErrEmptyTitle:
			logger.Error("empty document title", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrEmptyContent:
			logger.Error("empty document content", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrDatabaseError:
			logger.Error("failed to create document", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create document"})
		default:
			logger.Error("unexpected error while creating document", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	logger.Info("document created", "id", doc.ID)
	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) GetAllDocuments(c *gin.Context) {
	documents, err := h.service.GetAllDocuments(c.Request.Context())
	if err != nil {
		switch err {
		case domain.ErrDatabaseError:
			logger.Error("failed to get documents", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get documents"})
		default:
			logger.Error("unexpected error while getting documents", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	c.JSON(http.StatusOK, documents)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		logger.Error("invalid document id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrDocumentNotFound:
			logger.Warn("document not found", "id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		case domain.ErrDatabaseError:
			logger.Error("failed to get document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get document"})
		default:
			logger.Error("unexpected error while getting document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		logger.Error("invalid document id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	var doc domain.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		logger.Error("invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	doc.ID = id
	if err := h.service.UpdateDocument(c.Request.Context(), &doc); err != nil {
		switch err {
		case domain.ErrEmptyTitle:
			logger.Error("empty document title", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrEmptyContent:
			logger.Error("empty document content", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrDocumentNotFound:
			logger.Warn("document not found", "id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		case domain.ErrDatabaseError:
			logger.Error("failed to update document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update document"})
		default:
			logger.Error("unexpected error while updating document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		logger.Error("invalid document id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	if err := h.service.DeleteDocument(c.Request.Context(), id); err != nil {
		switch err {
		case domain.ErrDocumentNotFound:
			logger.Warn("document not found", "id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		case domain.ErrDatabaseError:
			logger.Error("failed to delete document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document"})
		default:
			logger.Error("unexpected error while deleting document", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
