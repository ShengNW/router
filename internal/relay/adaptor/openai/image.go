package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/internal/relay/model"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var imageResponse ImageResponse
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return wrapImageUpstreamError(resp, responseBody), nil
	}
	err = json.Unmarshal(responseBody, &imageResponse)
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, nil
}

func wrapImageUpstreamError(resp *http.Response, responseBody []byte) *model.ErrorWithStatusCode {
	var errorResponse struct {
		Error model.Error `json:"error"`
	}
	if err := json.Unmarshal(responseBody, &errorResponse); err == nil {
		if strings.TrimSpace(errorResponse.Error.Message) != "" {
			return &model.ErrorWithStatusCode{
				Error:      errorResponse.Error,
				StatusCode: resp.StatusCode,
			}
		}
	}
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message: fmt.Sprintf("upstream returned %s", resp.Status),
			Type:    "upstream_error",
			Code:    "upstream_http_error",
		},
		StatusCode: resp.StatusCode,
	}
}
