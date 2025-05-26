package models

type FileRecord struct {
	ID          string
	UserID      string
	StorageType string // "gridfs" or "local"
}

type FileResponse struct {
	ID string `json:"id"`
}

type FileMetadata struct {
	Size int64 `json:"size"`
}

type FileCreation struct {
	UserID      string
	StorageType string // "gridfs" or "local"
}
