package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tech-ip-sem2/shared/httpx"
)

type AuthClient struct {
	client  *httpx.Client
	baseURL string
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		client:  httpx.NewClient(baseURL, 3*time.Second),
		baseURL: baseURL,
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*VerifyResponse, int, error) {
	var resp VerifyResponse

	url := c.baseURL + "/v1/auth/verify"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	if requestID, ok := ctx.Value("requestID").(string); ok && requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	httpResp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, httpResp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return &resp, httpResp.StatusCode, nil
}

func (c *AuthClient) VerifyTokenWithHeader(ctx context.Context, token string) (*VerifyResponse, int, error) {
	return c.VerifyToken(ctx, token)
}
