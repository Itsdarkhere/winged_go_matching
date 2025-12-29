package integration

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
)

type BackendUploaderConfig struct {
	SupabaseUploaderConfig
}

type BackendStorage struct {
	*SupabaseUploader `json:"supabase_uploader_config" mapstructure:"supabase_uploader_config"`
}

type BackendUploader struct {
	*Storage[*BackendStorage]
}

func NewBackendUploader(cfg *BackendUploaderConfig) (*BackendUploader, error) {
	u, err := NewSupabaseUploader(&cfg.SupabaseUploaderConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating supabase uploader: %v", err)
	}

	return &BackendUploader{
		Storage: NewStorage[*BackendStorage](&BackendStorage{u}),
	}, nil
}

func (b *BackendStorage) PublicURL(ctx context.Context, fileName string) (string, error) {
	return b.SupabaseUploader.PublicURL(ctx, fileName)
}

func (b *BackendStorage) Upload(ctx context.Context, fileName string, file []byte) (*registration.FileDetails, error) {
	return b.SupabaseUploader.Upload(ctx, fileName, file)
}
