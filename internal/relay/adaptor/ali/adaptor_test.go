package ali

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	openaiadaptor "github.com/yeying-community/router/internal/relay/adaptor/openai"
	"github.com/yeying-community/router/internal/relay/meta"
	relaymodel "github.com/yeying-community/router/internal/relay/model"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

func TestGetRequestURL_ChatUsesCompatibleMode(t *testing.T) {
	adaptor := &Adaptor{}
	got, err := adaptor.GetRequestURL(&meta.Meta{
		Mode:    relaymode.ChatCompletions,
		BaseURL: "https://dashscope.aliyuncs.com",
	})
	if err != nil {
		t.Fatalf("GetRequestURL() error = %v", err)
	}
	want := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestGetRequestURL_ResponsesUsesAliCompatibleResponsesPath(t *testing.T) {
	adaptor := &Adaptor{}
	got, err := adaptor.GetRequestURL(&meta.Meta{
		Mode:    relaymode.Responses,
		BaseURL: "https://dashscope.aliyuncs.com",
	})
	if err != nil {
		t.Fatalf("GetRequestURL() error = %v", err)
	}
	want := "https://dashscope.aliyuncs.com/compatible-mode/v1/responses"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestGetRequestURL_QwenImageUsesMultimodalGenerationPath(t *testing.T) {
	adaptor := &Adaptor{}
	got, err := adaptor.GetRequestURL(&meta.Meta{
		Mode:            relaymode.ImagesGenerations,
		BaseURL:         "https://dashscope.aliyuncs.com",
		ActualModelName: "qwen-image-2.0-pro",
	})
	if err != nil {
		t.Fatalf("GetRequestURL() error = %v", err)
	}
	want := "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestGetRequestURL_LegacyImageUsesText2ImagePath(t *testing.T) {
	adaptor := &Adaptor{}
	got, err := adaptor.GetRequestURL(&meta.Meta{
		Mode:            relaymode.ImagesGenerations,
		BaseURL:         "https://dashscope.aliyuncs.com",
		ActualModelName: "wanx-v1",
	})
	if err != nil {
		t.Fatalf("GetRequestURL() error = %v", err)
	}
	want := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestGetRequestURL_QwenImageEditUsesMultimodalGenerationPath(t *testing.T) {
	adaptor := &Adaptor{}
	got, err := adaptor.GetRequestURL(&meta.Meta{
		Mode:            relaymode.ImagesEdits,
		BaseURL:         "https://dashscope.aliyuncs.com",
		ActualModelName: "qwen-image-2.0-pro",
		RequestURLPath:  "/v1/images/edits",
	})
	if err != nil {
		t.Fatalf("GetRequestURL() error = %v", err)
	}
	want := "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	if got != want {
		t.Fatalf("GetRequestURL() = %q, want %q", got, want)
	}
}

func TestConvertImageRequest_QwenImageUsesMultimodalMessages(t *testing.T) {
	adaptor := &Adaptor{}
	adaptor.Init(&meta.Meta{ActualModelName: "qwen-image-2.0"})

	converted, err := adaptor.ConvertImageRequest(&relaymodel.ImageRequest{
		Model:  "qwen-image-2.0",
		Prompt: "draw a blue square",
		Size:   "1024x1024",
	})
	if err != nil {
		t.Fatalf("ConvertImageRequest() error = %v", err)
	}
	qwenRequest, ok := converted.(*QwenImageRequest)
	if !ok {
		t.Fatalf("ConvertImageRequest() = %T, want *QwenImageRequest", converted)
	}
	if qwenRequest.Model != "qwen-image-2.0" {
		t.Fatalf("model = %q, want qwen-image-2.0", qwenRequest.Model)
	}
	if qwenRequest.Parameters.Size != "1024*1024" {
		t.Fatalf("size = %q, want 1024*1024", qwenRequest.Parameters.Size)
	}
	if len(qwenRequest.Input.Messages) != 1 || len(qwenRequest.Input.Messages[0].Content) != 1 || qwenRequest.Input.Messages[0].Content[0].Text != "draw a blue square" {
		t.Fatalf("messages = %#v, want prompt text content", qwenRequest.Input.Messages)
	}
}

func TestConvertQwenImageEditRequestUsesDataURIImageContent(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", "qwen-image-2.0"); err != nil {
		t.Fatalf("WriteField(model) error = %v", err)
	}
	if err := writer.WriteField("prompt", "make it blue"); err != nil {
		t.Fatalf("WriteField(prompt) error = %v", err)
	}
	part, err := writer.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("CreateFormFile(image) error = %v", err)
	}
	if _, err := part.Write([]byte{0x89, 0x50, 0x4e, 0x47}); err != nil {
		t.Fatalf("Write(image) error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	form, err := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary()).ReadForm(32 << 20)
	if err != nil {
		t.Fatalf("ReadForm() error = %v", err)
	}
	defer form.RemoveAll()

	converted, err := ConvertQwenImageEditRequest(relaymodel.ImageRequest{
		Model:  "qwen-image-2.0",
		Prompt: "make it blue",
		Size:   "1024x1024",
	}, form)
	if err != nil {
		t.Fatalf("ConvertQwenImageEditRequest() error = %v", err)
	}
	if converted.Parameters.Size != "1024*1024" {
		t.Fatalf("size = %q, want 1024*1024", converted.Parameters.Size)
	}
	if len(converted.Input.Messages) != 1 || len(converted.Input.Messages[0].Content) != 2 {
		t.Fatalf("messages = %#v, want image+text content", converted.Input.Messages)
	}
	if converted.Input.Messages[0].Content[0].Image == "" || converted.Input.Messages[0].Content[0].Text != "" {
		t.Fatalf("first content = %#v, want image data uri", converted.Input.Messages[0].Content[0])
	}
	if converted.Input.Messages[0].Content[1].Text != "make it blue" {
		t.Fatalf("second content = %#v, want prompt text", converted.Input.Messages[0].Content[1])
	}
}

func TestQwenImageHandlerWritesOpenAIImageResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

	body := []byte(`{"request_id":"req_123","output":{"choices":[{"message":{"content":[{"image":"https://example.com/image.png"}]}}]}}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	if err, _ := QwenImageHandler(ctx, resp); err != nil {
		t.Fatalf("QwenImageHandler() error = %+v", err)
	}

	var payload openaiadaptor.ImageResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal handler response failed: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].Url != "https://example.com/image.png" {
		t.Fatalf("handler payload data = %#v, want image url", payload.Data)
	}
	if count, ok := QwenImageOutputCount(ctx); !ok || count != 1 {
		t.Fatalf("QwenImageOutputCount() = (%d, %t), want (1, true)", count, ok)
	}
}

func TestResponseAli2OpenAIRetainsActualModelName(t *testing.T) {
	resp := responseAli2OpenAI(&ChatResponse{
		Error: Error{RequestId: "req_123"},
		Output: Output{
			Choices: []openaiadaptor.TextResponseChoice{
				{
					Message: relaymodel.Message{Content: "ok"},
				},
			},
		},
		Usage: Usage{InputTokens: 3, OutputTokens: 5},
	}, "qwen-plus-latest")

	if resp.Model != "qwen-plus-latest" {
		t.Fatalf("responseAli2OpenAI().Model = %q, want %q", resp.Model, "qwen-plus-latest")
	}
}

func TestStreamResponseAli2OpenAIRetainsActualModelName(t *testing.T) {
	resp := streamResponseAli2OpenAI(&ChatResponse{
		Error: Error{RequestId: "req_123"},
		Output: Output{
			Choices: []openaiadaptor.TextResponseChoice{
				{
					Message:      relaymodel.Message{Content: "delta"},
					FinishReason: "stop",
				},
			},
		},
	}, "qwen-max-latest")

	if resp == nil {
		t.Fatal("streamResponseAli2OpenAI() returned nil")
	}
	if resp.Model != "qwen-max-latest" {
		t.Fatalf("streamResponseAli2OpenAI().Model = %q, want %q", resp.Model, "qwen-max-latest")
	}
}

func TestHandlerWritesActualModelName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	body := []byte(`{"request_id":"req_123","output":{"choices":[{"message":{"content":"ok"}}]},"usage":{"input_tokens":2,"output_tokens":4}}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")

	if err, _ := Handler(ctx, resp, "qwen-turbo-latest"); err != nil {
		t.Fatalf("Handler() error = %+v", err)
	}

	var payload openaiadaptor.TextResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal handler response failed: %v", err)
	}
	if payload.Model != "qwen-turbo-latest" {
		t.Fatalf("handler payload model = %q, want %q", payload.Model, "qwen-turbo-latest")
	}
}
