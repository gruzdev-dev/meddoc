package file

import "mime/multipart"

// FileHeaderAdapter адаптирует *multipart.FileHeader к интерфейсу FileOpener
type FileHeaderAdapter struct {
	*multipart.FileHeader
}

func (a *FileHeaderAdapter) GetFilename() string {
	return a.Filename
}

func (a *FileHeaderAdapter) GetHeader() map[string][]string {
	return a.Header
}

func (a *FileHeaderAdapter) GetSize() int64 {
	return a.Size
}

func (a *FileHeaderAdapter) Open() (multipart.File, error) {
	return a.FileHeader.Open()
}
