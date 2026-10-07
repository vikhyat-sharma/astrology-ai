package mocks

import (
	"io"
	"net/http"
)

// MockHTTPClient is a test double for ports.HTTPClient.
// Set PostFunc to control what the mock returns for any given call.
type MockHTTPClient struct {
	PostFunc func(url, contentType string, body io.Reader) (*http.Response, error)
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if m.PostFunc != nil {
		return m.PostFunc(url, contentType, body)
	}
	return nil, nil
}
