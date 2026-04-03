package mocks

import (
	"io"
	"net/http"
)

// MockHTTPClient is a mock implementation of interfaces.HTTPClientInterface
type MockHTTPClient struct {
	PostFunc func(url, contentType string, body io.Reader) (*http.Response, error)
}

// Post mocks the Post method
func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if m.PostFunc != nil {
		return m.PostFunc(url, contentType, body)
	}
	return nil, nil
}
