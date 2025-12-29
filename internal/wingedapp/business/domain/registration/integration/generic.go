package integration

import (
	"context"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
)

// Uploader is an interface for uploading files and getting their public URLs.
type uploader interface {
	Upload(ctx context.Context, fileName string, file []byte) (*registration.FileDetails, error)
	PublicURL(ctx context.Context, fileName string) (string, error)
}

type Storage[T uploader] struct {
	T T
}

func (s *Storage[T]) Upload(ctx context.Context, fileName string, file []byte) (*registration.FileDetails, error) {
	return s.T.Upload(ctx, fileName, file)
}

func (s *Storage[T]) PublicURL(ctx context.Context, fileName string) (string, error) {
	return s.T.PublicURL(ctx, fileName)
}

func NewStorage[T uploader](uploader T) *Storage[T] {
	return &Storage[T]{T: uploader}
}
