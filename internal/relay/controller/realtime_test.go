package controller

import (
	"net/http"
	"testing"

	"github.com/yeying-community/router/internal/relay/meta"
)

func TestNormalizeRealtimeWebSocketURL(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "https://api.openai.com/v1/realtime?model=gpt-realtime-2", want: "wss://api.openai.com/v1/realtime?model=gpt-realtime-2"},
		{raw: "http://localhost:8080/v1/realtime", want: "ws://localhost:8080/v1/realtime"},
	}
	for _, tt := range tests {
		got, err := normalizeRealtimeWebSocketURL(tt.raw)
		if err != nil {
			t.Fatalf("normalizeRealtimeWebSocketURL(%q) returned error: %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("normalizeRealtimeWebSocketURL(%q)=%q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestRealtimeUpgradeHeadersCopiesSubprotocol(t *testing.T) {
	header := realtimeUpgradeHeaders(nil)
	if got := header.Get("Sec-WebSocket-Protocol"); got != "" {
		t.Fatalf("Sec-WebSocket-Protocol = %q, want empty", got)
	}
}

func TestCloneRealtimeRequestHeadersDropsHopByHop(t *testing.T) {
	header := http.Header{
		"Authorization":          []string{"Bearer sk-test"},
		"OpenAI-Beta":            []string{"realtime=v1"},
		"Connection":             []string{"Upgrade"},
		"Sec-WebSocket-Key":      []string{"secret"},
		"Sec-WebSocket-Version":  []string{"13"},
		"Sec-WebSocket-Protocol": []string{"realtime"},
	}
	cloned := cloneRealtimeRequestHeaders(header, &meta.Meta{APIKey: "upstream-key"})
	if cloned.Get("Authorization") != "Bearer upstream-key" {
		t.Fatalf("Authorization = %q, want Bearer upstream-key", cloned.Get("Authorization"))
	}
	if cloned.Get("OpenAI-Beta") != "realtime=v1" {
		t.Fatalf("OpenAI-Beta = %q, want realtime=v1", cloned.Get("OpenAI-Beta"))
	}
	if cloned.Get("Connection") != "" {
		t.Fatalf("Connection = %q, want empty", cloned.Get("Connection"))
	}
	if cloned.Get("Sec-WebSocket-Key") != "" {
		t.Fatalf("Sec-WebSocket-Key = %q, want empty", cloned.Get("Sec-WebSocket-Key"))
	}
	if cloned.Get("Sec-WebSocket-Protocol") != "realtime" {
		t.Fatalf("Sec-WebSocket-Protocol = %q, want realtime", cloned.Get("Sec-WebSocket-Protocol"))
	}
}
