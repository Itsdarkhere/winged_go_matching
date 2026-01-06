package extmatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/hashicorp/go-retryablehttp"
)

type QualitativeMatcher struct {
	cln *HttpClient
}

func NewQualitativeMatcher(logger applog.Logger) *QualitativeMatcher {
	cln := NewHttpClient(&HttpClientCfg{
		GeneralTimeoutDuration: 120,
		ReadTimeoutDuration:    120,
		RetryDuration:          60,
		MaxRetries:             3,
	}, logger)

	return &QualitativeMatcher{cln}
}

func (qm *QualitativeMatcher) Qualify(
	ctx context.Context,
	req *matching.QualitativeMatchRequest,
) (*matching.MatchCompatibilityResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate qualitative match request: %w", err)
	}

	// TODO: switch URL based on environment
	url := "https://winged-ai-backend-dev-1020977593789.asia-east1.run.app/matchmaking_v2"

	ctx, cancel := context.WithTimeout(ctx, time.Duration(120)*time.Second)
	defer cancel()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request data: %w", err)
	}

	fmt.Println("==== AI BACKEND REQUEST (before send) ====")
	fmt.Println(string(body))
	fmt.Println("===========================================")

	reqCtx, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	reqCtx.Header = make(map[string][]string)
	reqCtx.Header["X-API-Key"] = []string{"winged-api-key"}

	resp, err := qm.cln.Client.Do(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	respBody, err := readResponseBody(ctx, resp)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	fmt.Println("==== AI BACKEND REQUEST ====")
	fmt.Println(string(body))
	fmt.Println("==== AI BACKEND RESPONSE ====")
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", string(respBody))
	fmt.Println("============================")

	var compatRes matching.MatchCompatibilityResult
	if err := json.Unmarshal(respBody, &compatRes); err != nil {
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &compatRes, nil
}
