package middleware

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRequestModel_VideosMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", "veo-3.0-generate-preview"); err != nil {
		t.Fatalf("WriteField(model) error: %v", err)
	}
	if err := writer.WriteField("prompt", "test"); err != nil {
		t.Fatalf("WriteField(prompt) error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error: %v", err)
	}

	req := httptest.NewRequest("POST", "/v1/videos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	modelName, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("getRequestModel returned error: %v", err)
	}
	if modelName != "veo-3.0-generate-preview" {
		t.Fatalf("getRequestModel returned %q, want %q", modelName, "veo-3.0-generate-preview")
	}
}

func TestGetRequestModel_ImageEditsMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", "qwen-image-2.0"); err != nil {
		t.Fatalf("WriteField(model) error: %v", err)
	}
	if err := writer.WriteField("prompt", "make it blue"); err != nil {
		t.Fatalf("WriteField(prompt) error: %v", err)
	}
	part, err := writer.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("CreateFormFile(image) error: %v", err)
	}
	if _, err := part.Write([]byte{0x89, 0x50, 0x4e, 0x47}); err != nil {
		t.Fatalf("Write(image) error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error: %v", err)
	}

	req := httptest.NewRequest("POST", "/v1/images/edits", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	modelName, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("getRequestModel returned error: %v", err)
	}
	if modelName != "qwen-image-2.0" {
		t.Fatalf("getRequestModel returned %q, want %q", modelName, "qwen-image-2.0")
	}
}

func TestGetRequestModel_VideoStatusQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/v1/videos/task_123?model=veo-3.0-generate-preview", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	modelName, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("getRequestModel returned error: %v", err)
	}
	if modelName != "veo-3.0-generate-preview" {
		t.Fatalf("getRequestModel returned %q, want %q", modelName, "veo-3.0-generate-preview")
	}
}

func TestGetRequestModel_RealtimeQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("POST", "/v1/realtime/calls?model=gpt-realtime-2", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	modelName, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("getRequestModel returned error: %v", err)
	}
	if modelName != "gpt-realtime-2" {
		t.Fatalf("getRequestModel returned %q, want %q", modelName, "gpt-realtime-2")
	}
}

func TestGetRequestModel_RealtimeNestedSessionModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("POST", "/v1/realtime/client_secrets", bytes.NewBufferString(`{"session":{"model":"gpt-realtime-1.5"}}`))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	modelName, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("getRequestModel returned error: %v", err)
	}
	if modelName != "gpt-realtime-1.5" {
		t.Fatalf("getRequestModel returned %q, want %q", modelName, "gpt-realtime-1.5")
	}
}
