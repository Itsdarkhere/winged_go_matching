package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/auth"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

/*
	This is the Supabase uploader, and here are some architectural designs of relevance:

	The Supabase client is not stored in the BackendUploader struct to avoid re-authing in Supabase.
	Instead, the Supabase client is extracted from the context in each method.
	These clients come from a previous middleware execution th
*/

var (
	ErrAuthCtxIsMissing  = errors.New("supabase auth context is missing")
	ErrSupabaseClientNil = errors.New("supabase client is nil")
)

// SupabaseUploader is our uploader implementation for fileuploader.
type SupabaseUploader struct {
	cfg *SupabaseUploaderConfig
}

func (s *SupabaseUploaderConfig) Validate() error {
	return validationlib.Validate(s)
}

// NewSupabaseUploader creates a new SupabaseUploader instance.
func NewSupabaseUploader(cfg *SupabaseUploaderConfig) (*SupabaseUploader, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	sBaseUploader := &SupabaseUploader{
		cfg: cfg,
	}

	return sBaseUploader, nil
}

type SupabaseUploaderConfig struct {
	Bucket     string `json:"bucket" validate:"required"`
	AnonPubKey string `json:"anon_pub_key" mapstructure:"anon_pub_key" validate:"required"`
	ProjID     string `json:"proj_id" mapstructure:"proj_id" validate:"required"`
}

// supabaseClient extracts the supabase client from context.
func supabaseClient(ctx context.Context) (*supabase.Client, error) {
	authedCln, ok := ctx.Value(auth.CtxSupabaseCln).(*supabase.Client)
	if !ok {
		return nil, ErrAuthCtxIsMissing
	}
	if authedCln == nil {
		return nil, ErrSupabaseClientNil
	}
	return authedCln, nil
}

func (s *SupabaseUploader) Upload(ctx context.Context, fileName string, file []byte) (*registration.FileDetails, error) {
	authedCln, err := supabaseClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("supabase client: %w", err)
	}

	contentType := http.DetectContentType(file)
	upsert := true
	fileOpts := storage_go.FileOptions{
		ContentType: &contentType,
		Upsert:      &upsert,
	}
	uploadFileResp, err := authedCln.Storage.UploadFile(s.cfg.Bucket, fileName, bytes.NewReader(file), fileOpts)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	return fileUploadResToFileDetails(&uploadFileResp), nil
}

func fileUploadResToFileDetails(resp *storage_go.FileUploadResponse) *registration.FileDetails {
	return &registration.FileDetails{
		Bucket: resp.Key,
		Key:    resp.Key,
	}
}

func (s *SupabaseUploader) PublicURL(ctx context.Context, fileName string) (string, error) {
	authedCln, err := supabaseClient(ctx)
	if err != nil {
		return "", fmt.Errorf("supabase client: %w", err)
	}

	const (
		// TODO: make this config driven
		expiresIn = 60 * 60 // 1 hour (60 "60 seconds")
	)
	signedURL, err := authedCln.Storage.CreateSignedUrl(s.cfg.Bucket, fileName, expiresIn)
	if err != nil {
		return "", fmt.Errorf("create signed url: %w", err)
	}

	return signedURL.SignedURL, nil
}
