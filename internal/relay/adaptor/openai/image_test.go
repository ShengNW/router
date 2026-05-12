package openai

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestImageHandlerWrapsNonJSONUpstreamError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Status:     "502 Bad Gateway",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("<html>bad gateway</html>")),
	}

	err, usage := ImageHandler(nil, resp)
	if usage != nil {
		t.Fatalf("usage = %+v, want nil", usage)
	}
	if err == nil {
		t.Fatal("ImageHandler() error = nil, want upstream error")
	}
	if err.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want %d", err.StatusCode, http.StatusBadGateway)
	}
	if err.Type != "upstream_error" {
		t.Fatalf("Type = %q, want upstream_error", err.Type)
	}
	if err.Code != "upstream_http_error" {
		t.Fatalf("Code = %v, want upstream_http_error", err.Code)
	}
	if err.Message != "upstream returned 502 Bad Gateway" {
		t.Fatalf("Message = %q", err.Message)
	}
}
