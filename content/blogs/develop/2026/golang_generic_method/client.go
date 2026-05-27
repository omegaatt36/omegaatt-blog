package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string, client *http.Client) *Client {
	return &Client{baseURL: baseURL, client: client}
}

func doRequest[T any](ctx context.Context, client *Client, method, path string, body any) (*T, error) {
	var response Response[T]
	err := client.doRequest(ctx, method, path, body, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func doRequestWithQuery[T any](ctx context.Context, client *Client, method, path string, query url.Values, body any) (*T, error) {
	var response Response[T]
	err := client.doRequestWithQuery(ctx, method, path, query, body, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	return c.doRequestWithQuery(ctx, method, path, nil, body, result)
}

func (c *Client) doRequestWithQuery(ctx context.Context, method, requestPath string, query url.Values, body any, result any) error {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	baseURL.Path = path.Join(baseURL.Path, requestPath)

	if len(query) > 0 {
		baseURL.RawQuery = query.Encode()
	}

	u := baseURL.String()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var errorResponse Response[any]
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil && errorResponse.Error != "" {
			return fmt.Errorf("API error: %s", errorResponse.Error)
		}

		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	if len(bytes.TrimSpace(responseBody)) == 0 || result == nil {
		return nil
	}

	if err := json.Unmarshal(responseBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

type Response[T any] struct {
	Data  T      `json:"data,omitzero"`
	Error string `json:"error,omitzero"`
}

type RequestGetResourceA struct {
	ID    string
	Query string
}

type ResponseGetResourceA struct {
	ID string
}

type responseGetResourceA struct {
	ID string `json:"id"`
}

type APIClient struct {
	client *Client
}

func (a *APIClient) GetResourceA(ctx context.Context, req RequestGetResourceA) (*ResponseGetResourceA, error) {
	path := fmt.Sprintf("/api/v1/a/%s", req.ID)

	response, err := doRequest[responseGetResourceA](ctx, a.client, http.MethodGet, path, struct {
		Query string `json:"query"`
	}{
		Query: req.Query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get scan metrics for agent '%s': %w", req.ID, err)
	}

	return &ResponseGetResourceA{
		ID: response.ID,
	}, nil
}

type RequestGetResourceB struct {
	ID    string
	Query string
}

type ResponseGetResourceB struct {
	ID string
}

type responseGetResourceB struct {
	ID string `json:"id"`
}

func (a *APIClient) GetResourceB(ctx context.Context, req RequestGetResourceB) (*ResponseGetResourceB, error) {
	path := fmt.Sprintf("/api/v1/b/%s", req.ID)

	response, err := doRequest[responseGetResourceB](ctx, a.client, http.MethodGet, path, struct {
		Query string `json:"query"`
	}{
		Query: req.Query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource b for '%s': %w", req.ID, err)
	}

	return &ResponseGetResourceB{
		ID: response.ID,
	}, nil
}
