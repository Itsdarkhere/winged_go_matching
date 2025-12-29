package extmatcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"

	"github.com/hashicorp/go-retryablehttp"
)

type HttpClient struct {
	Client *retryablehttp.Client
	cfg    *HttpClientCfg
	logger applog.Logger
}

type HttpClientCfg struct {
	GeneralTimeoutDuration int `json:"CLIENT_GENERAL_TIMEOUT_DURATION" mapstructure:"CLIENT_GENERAL_TIMEOUT_DURATION" validate:"required"`
	ReadTimeoutDuration    int `json:"CLIENT_READ_TIMEOUT_DURATION" mapstructure:"CLIENT_READ_TIMEOUT_DURATION" validate:"required"`
	RetryDuration          int `json:"CLIENT_RETRY_TIMEOUT_DURATION" mapstructure:"CLIENT_RETRY_TIMEOUT_DURATION" validate:"required"`
	MaxRetries             int `json:"CLIENT_MAX_RETRIES" mapstructure:"CLIENT_MAX_RETRIES" validate:"required"`
}

func (h *HttpClientCfg) Validate() error {
	return validationlib.Validate(h)
}

func newRetryableHttpClient(cfg *HttpClientCfg, logger applog.Logger) *retryablehttp.Client {
	r := retryablehttp.NewClient()
	r.RetryMax = cfg.MaxRetries
	r.RetryWaitMin = time.Duration(cfg.RetryDuration) * time.Millisecond
	r.Backoff = retryablehttp.DefaultBackoff
	r.Logger = applog.NewLeveledLogger(logger)
	r.HTTPClient.Timeout = time.Duration(cfg.GeneralTimeoutDuration) * time.Second
	return r
}

func NewHttpClient(cfg *HttpClientCfg, logger applog.Logger) *HttpClient {
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("validate HttpClient: %v", err))
	}
	if logger == nil {
		panic("logger is required")
	}

	return &HttpClient{
		Client: newRetryableHttpClient(cfg, logger),
		cfg:    cfg,
		logger: logger,
	}
}

func readResponseBody(ctx context.Context, resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// Get performs a GET request to the specified URL with a timeout.
func (h *HttpClient) Get(ctx context.Context, url string) ([]byte, *http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(h.cfg.GeneralTimeoutDuration)*time.Second)
	defer cancel()

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("doReq: %w", err)
	}

	body, err := readResponseBody(ctx, resp)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, errors.New("not found")
	}

	return body, resp, nil
}
