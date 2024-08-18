// Package anki provides a client for interacting with the Anki desktop application's API.
package anki

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Default values for the Anki API
const (
	defaultBaseURL = "http://localhost:8765"
	apiVersion     = 6
)

// Client represents an Anki API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	Notes      *NoteService
	ModelNames *ModelNameService
	DeckNames  *DeckNameService
}

// requestResult represents the structure of the Anki API response.
type requestResult struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

// NewClient creates a new Anki API client with default settings.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	c.Notes = NewNoteService(c)
	c.ModelNames = NewModelNameService(c)
	c.DeckNames = NewDeckNameService(c)
	return c
}

// SetHTTPClient sets a custom HTTP client for the Anki API client.
func (c *Client) SetHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// SetBaseURL sets a custom base URL for the Anki API client.
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// SetTimeout sets a custom timeout for the HTTP client.
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// BaseURL returns the current base URL of the Anki API client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// HTTPClient returns the current HTTP client used by the Anki API client.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// Close closes idle connections in the HTTP client.
func (c *Client) Close() {
	if c.httpClient == nil {
		return
	}
	c.httpClient.CloseIdleConnections()
}

type ClientResponse struct {
	Result  json.RawMessage `json:"result"`
	Payload []byte          `json:"payload"`
}

// Send sends a request to the Anki API and returns the raw JSON response.
func (c *Client) Send(action string, params interface{}) (ClientResponse, error) {
	payload, err := NewRequestPayload(action, params)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("creating payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return ClientResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	var result requestResult
	if err = json.Unmarshal(body, &result); err != nil {
		return ClientResponse{}, fmt.Errorf("unmarshaling response: %w", err)
	}
	if result.Error != nil {
		return ClientResponse{}, errors.New(*result.Error)
	}

	return ClientResponse{
		Result:  result.Result,
		Payload: payload,
	}, nil
}

// sendAndUnmarshal sends a request to the Anki API and unmarshals the response into the provided interface.
func (c *Client) sendAndUnmarshal(action string, params, v interface{}) error {
	result, err := c.Send(action, params)
	if err != nil {
		return err
	}
	return json.Unmarshal(result.Result, v)
}

// NewRequestPayload creates a new payload for the Anki API request.
func NewRequestPayload(action string, params interface{}) ([]byte, error) {
	payload := map[string]interface{}{
		"action":  action,
		"version": apiVersion,
	}
	if params != nil {
		payload["params"] = params
	}
	return json.Marshal(payload)
}
