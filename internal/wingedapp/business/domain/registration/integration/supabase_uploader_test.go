package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"wingedapp/pgtester/internal/testhelper"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration/integration"
	"wingedapp/pgtester/internal/wingedapp/testsuite"
	"wingedapp/pgtester/internal/wingedapp/testsuite/asset"

	"github.com/stretchr/testify/require"
)

func TestSupabaseUploader_Upload(t *testing.T) {
	t.Skip() // skip for now cause speed
	tSuite := testsuite.New(t)

	bucket, cleanup := tSuite.CreateSupabaseRandomBucket()
	defer cleanup()

	cfg := &integration.SupabaseUploaderConfig{
		Bucket:     bucket.Name,
		AnonPubKey: tSuite.Config.SupabaseAnonPublicKey,
		ProjID:     tSuite.Config.SupabaseProjID,
	}

	uploader, err := integration.NewSupabaseUploader(cfg)
	require.NoError(t, err, "expected no error when creating BackendUploader")
	require.NotNil(t, uploader, "expected non-nil BackendUploader instance")

	ctxWSupabaseCln := context.WithValue(context.Background(), "supabase_cln", tSuite.AuthedSupabaseClient())

	testMultipartFile := testhelper.MockMultipartFileHeaders(t)[0]
	f, err := testMultipartFile.Open()
	require.NoError(t, err, "expected no error opening multipart file")

	fBytes, err := io.ReadAll(f)
	require.NoError(t, err, "expected no error reading multipart file")
	require.NotNil(t, fBytes, "expected multipart file to not be nil")

	resp, err := uploader.Upload(ctxWSupabaseCln,
		"testfile.txt",
		fBytes,
	)
	require.NoError(t, err, "expected no error when uploading multipart file")
	require.NotNil(t, resp, "expected non-nil response from upload")
}

func TestSupabaseUploader_Upload_FailNoSetCtx(t *testing.T) {
	cfg := &integration.SupabaseUploaderConfig{
		Bucket:     "n/a",
		AnonPubKey: "n/a",
		ProjID:     "n/a",
	}
	uploader, err := integration.NewSupabaseUploader(cfg)
	require.NoError(t, err, "expected no error when creating BackendUploader")
	require.NotNil(t, uploader, "expected non-nil BackendUploader instance")

	testMultipartFile := testhelper.MockMultipartFileHeaders(t)[0]
	f, err := testMultipartFile.Open()
	require.NoError(t, err, "expected no error opening multipart file")

	fBytes, err := io.ReadAll(f)
	require.NoError(t, err, "expected no error reading multipart file")
	require.NotNil(t, fBytes, "expected multipart file to not be nil")

	_, err = uploader.Upload(context.Background(),
		"testfile.txt",
		fBytes,
	)
	require.Error(t, err, "expected error because of missing supabase client in context")
}

func TestCleanTestBuckets(t *testing.T) {
	t.Skip()
	tSuite := testsuite.New(t)
	tSuite.CleanTestBuckets()
}

func TestUploadPNG(t *testing.T) {
	t.Skip() // might need to restore later, for now.. just skip
	tSuite := testsuite.New(t)

	bucket, _ := tSuite.CreateSupabaseRandomBucket()
	authedCln := tSuite.AuthedSupabaseClient()
	f, err := asset.ImageFile()
	require.NoError(t, err, "expected no error getting dolphin png asset")

	fBytes, err := io.ReadAll(f)
	require.NoError(t, err, "expected no error reading dolphin png asset")
	require.NotNil(t, fBytes, "expected non-nil bytes from dolphin png asset")

	res, err := authedCln.Storage.UploadFile(bucket.Name, "image.png", bytes.NewReader(fBytes))
	require.NoError(t, err, "expected no error when uploading to supabase bucket")
	fmt.Println("=== res:", res)

	presignedURL, err := authedCln.Storage.CreateSignedUrl(bucket.Name, "image.png", 60)
	require.NoError(t, err, "expected no error when creating presigned URL")
	fmt.Println("=== presignedURL:", presignedURL)
}
