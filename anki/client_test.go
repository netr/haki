package anki_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/netr/haki/anki"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{"Default URL", "", "http://localhost:8765"},
		{"Custom URL", "http://custom:8080", "http://custom:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := anki.NewClient(tt.baseURL)
			if client.BaseURL() != tt.expected {
				t.Errorf("NewClient(%q) baseURL = %q, want %q", tt.baseURL, client.BaseURL(), tt.expected)
			}
			if client.HTTPClient() == nil {
				t.Error("NewClient() httpClient is nil")
			}
			if client.Notes == nil {
				t.Error("NewClient() Notes service is nil")
			}
			if client.ModelNames == nil {
				t.Error("NewClient() ModelNames service is nil")
			}
			if client.DeckNames == nil {
				t.Error("NewClient() DeckNames service is nil")
			}
		})
	}
}

func TestClient_SetHTTPClient(t *testing.T) {
	client := anki.NewClient("")
	customHTTPClient := &http.Client{Timeout: 5 * time.Second}

	client.SetHTTPClient(customHTTPClient)

	if client.HTTPClient() != customHTTPClient {
		t.Error("SetHTTPClient() did not set the custom HTTP client")
	}
}

func TestClient_SetBaseURL(t *testing.T) {
	client := anki.NewClient("")
	newBaseURL := "http://newbase:8080"

	client.SetBaseURL(newBaseURL)

	if client.BaseURL() != newBaseURL {
		t.Errorf("SetBaseURL(%q) did not set the new base URL", newBaseURL)
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := anki.NewClient("")
	newTimeout := 15 * time.Second

	client.SetTimeout(newTimeout)

	if client.HTTPClient().Timeout != newTimeout {
		t.Errorf("SetTimeout(%v) did not set the new timeout", newTimeout)
	}
}

func TestClient_Send(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		expectedError  bool
	}{
		{"Successful response", `{"result": {"key": "value"}}`, false},
		{"Error response", `{"error": "An error occurred"}`, true},
		{"Invalid JSON", `{invalid json}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(tt.serverResponse))
				if err != nil {
					t.Fatalf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := anki.NewClient(server.URL)
			result, err := client.Send("test_action", nil)

			if tt.expectedError && err == nil {
				t.Error("Expected an error, but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectedError && result.Result == nil {
				t.Error("Expected a result, but got nil")
			}
		})
	}
}

func TestNewRequestPayload(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		params   interface{}
		expected string
	}{
		{"Action only", "test_action", nil, `{"action":"test_action","version":6}`},
		{"Action with params", "test_action", map[string]string{"key": "value"}, `{"action":"test_action","params":{"key":"value"},"version":6}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := anki.NewRequestPayload(tt.action, tt.params)
			if err != nil {
				t.Fatalf("NewRequestPayload() returned an error: %v", err)
			}

			var result map[string]interface{}
			err = json.Unmarshal(payload, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			expectedResult := make(map[string]interface{})
			err = json.Unmarshal([]byte(tt.expected), &expectedResult)
			if err != nil {
				t.Fatalf("Failed to unmarshal expected result: %v", err)
			}

			if !jsonEqual(result, expectedResult) {
				t.Errorf("NewRequestPayload() = %v, want %v", string(payload), tt.expected)
			}
		})
	}
}

// Helper function to compare JSON objects
func jsonEqual(a, b map[string]interface{}) bool {
	return string(mustMarshal(a)) == string(mustMarshal(b))
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
