package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	relaychannel "github.com/yeying-community/router/internal/relay/channel"
	"github.com/yeying-community/router/internal/relay/meta"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

func TestSetupRequestHeaderSetsJSONAcceptForNonStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Request.Header.Set("Accept", "text/event-stream")
	ctx.Request.Header.Set("Content-Type", "application/json")

	req := httptest.NewRequest(http.MethodPost, "https://example.com/v1/chat/completions", nil)
	adaptor := &Adaptor{ChannelProtocol: relaychannel.OpenAI}
	meta := &meta.Meta{APIKey: "sk-test", ChannelProtocol: relaychannel.OpenAI}

	if err := adaptor.SetupRequestHeader(ctx, req, meta); err != nil {
		t.Fatalf("SetupRequestHeader returned error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q, want %q", got, "application/json")
	}
}

func TestSetupRequestHeaderSetsSSEAcceptForStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Request.Header.Set("Accept", "application/json")
	ctx.Request.Header.Set("Content-Type", "application/json")

	req := httptest.NewRequest(http.MethodPost, "https://example.com/v1/chat/completions", nil)
	adaptor := &Adaptor{ChannelProtocol: relaychannel.OpenAI}
	meta := &meta.Meta{
		APIKey:          "sk-test",
		ChannelProtocol: relaychannel.OpenAI,
		IsStream:        true,
	}

	if err := adaptor.SetupRequestHeader(ctx, req, meta); err != nil {
		t.Fatalf("SetupRequestHeader returned error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "text/event-stream" {
		t.Fatalf("Accept = %q, want %q", got, "text/event-stream")
	}
}

func TestSetupRequestHeaderPreservesAudioAcceptForSpeech(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/audio/speech", nil)
	ctx.Request.Header.Set("Accept", "audio/mpeg")
	ctx.Request.Header.Set("Content-Type", "application/json")

	req := httptest.NewRequest(http.MethodPost, "https://example.com/v1/audio/speech", nil)
	adaptor := &Adaptor{ChannelProtocol: relaychannel.OpenAI}
	meta := &meta.Meta{
		APIKey:          "sk-test",
		ChannelProtocol: relaychannel.OpenAI,
		Mode:            relaymode.AudioSpeech,
	}

	if err := adaptor.SetupRequestHeader(ctx, req, meta); err != nil {
		t.Fatalf("SetupRequestHeader returned error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "audio/mpeg" {
		t.Fatalf("Accept = %q, want %q", got, "audio/mpeg")
	}
}

func TestDoResponseRelaysRawResponseForRealtime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/realtime/client_secrets", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"client_secret":"secret_123"}`)),
	}
	adaptor := &Adaptor{ChannelProtocol: relaychannel.OpenAI}
	meta := &meta.Meta{
		ChannelProtocol: relaychannel.OpenAI,
		Mode:            relaymode.Realtime,
	}

	usage, err := adaptor.DoResponse(ctx, resp, meta)
	if err != nil {
		t.Fatalf("DoResponse returned error: %v", err)
	}
	if usage != nil {
		t.Fatalf("usage = %#v, want nil", usage)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != `{"client_secret":"secret_123"}` {
		t.Fatalf("body = %q, want raw passthrough", body)
	}
}
