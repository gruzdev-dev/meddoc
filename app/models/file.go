package models

type File struct {
	ID          string `json:"id" binding:"required"`
	UserID      string `json:"-" binding:"required"`
	DownloadURL string `json:"download_url" binding:"required"`
	StorageType string `json:"-" binding:"required"` // "gridfs" or "local"
}
