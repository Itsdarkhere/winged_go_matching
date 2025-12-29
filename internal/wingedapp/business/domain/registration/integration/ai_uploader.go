package integration

import (
	"context"
	"fmt"
	registration "wingedapp/pgtester/internal/wingedapp/business/domain/registration"
)

type AIUploader struct {
	*Storage[*AIStorage]
}

type AIUploaderConfig struct {
	SupabaseUploaderConfig `json:"supabase_uploader_config" mapstructure:"supabase_uploader_config"`
}

func NewAIUploader(cfg *AIUploaderConfig) (*AIUploader, error) {
	u, err := NewSupabaseUploader(&cfg.SupabaseUploaderConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating supabase uploader: %v", err)
	}

	return &AIUploader{
		Storage: NewStorage[*AIStorage](&AIStorage{u}),
	}, nil
}

type AIStorage struct {
	*SupabaseUploader
}

func (a *AIStorage) PublicURL(ctx context.Context, fileName string) (string, error) {
	// This is where we have to be careful
	return a.SupabaseUploader.PublicURL(ctx, fileName)
}

func (a *AIStorage) Upload(ctx context.Context, fileName string, file []byte) (*registration.FileDetails, error) {
	return a.SupabaseUploader.Upload(ctx, fileName, file)
}
