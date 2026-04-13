package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yeying-community/router/common/helper"
)

var errorCardWriteMu sync.Mutex

type ErrorCardAction struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
	Mode  string `json:"mode,omitempty"`
	URL   string `json:"url,omitempty"`
}

type ErrorCardEvent struct {
	EventType string `json:"event_type"`
	Domain    string `json:"domain"`
	Subtype   string `json:"subtype"`
	Severity  string `json:"severity"`

	Title        string `json:"title"`
	Summary      string `json:"summary"`
	BizStatus    string `json:"biz_status"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`

	ImpactScope   string `json:"impact_scope"`
	ImpactSummary string `json:"impact_summary"`

	UserID    string `json:"user_id,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	ModelName string `json:"model_name,omitempty"`

	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Occurred  string `json:"occurred_at"`

	HTTPStatus     int    `json:"http_status,omitempty"`
	UpstreamStatus int    `json:"upstream_status,omitempty"`
	RetryCount     int    `json:"retry_count,omitempty"`
	ProviderStatus string `json:"provider_status,omitempty"`
	ProviderCode   string `json:"provider_code,omitempty"`

	Service   string `json:"service,omitempty"`
	DetailURL string `json:"detail_url,omitempty"`
	Runbook   string `json:"runbook_url,omitempty"`

	Tags    map[string]string `json:"tags,omitempty"`
	Actions []ErrorCardAction `json:"actions,omitempty"`
}

// EmitFeishuCardError writes a single-line Feishu interactive-card JSON payload
// into logs/error.log, with normalized event fields attached for machine parsing.
func EmitFeishuCardError(ctx context.Context, event ErrorCardEvent) {
	SetupLogger()
	normalized := normalizeErrorCardEvent(ctx, event)
	line, err := buildErrorCardLine(normalized)
	if err != nil {
		logHelper(ctx, loggerError, "[error_card] marshal_failed: "+err.Error())
		return
	}
	writeErrorCardLine(line)
}

func normalizeErrorCardEvent(ctx context.Context, event ErrorCardEvent) ErrorCardEvent {
	event.Domain = normalizeDomain(event.Domain)
	event.Subtype = normalizeDefault(strings.TrimSpace(event.Subtype), "unknown")
	event.Severity = normalizeSeverity(event.Severity)
	event.BizStatus = normalizeDefault(strings.TrimSpace(event.BizStatus), "failed")
	event.ErrorCode = normalizeDefault(strings.TrimSpace(event.ErrorCode), "UNKNOWN")
	if event.TraceID == "" && ctx != nil {
		event.TraceID = strings.TrimSpace(helper.GetTraceID(ctx))
	}
	event.RequestID = normalizeDefault(strings.TrimSpace(event.RequestID), event.TraceID)
	event.Occurred = normalizeDefault(strings.TrimSpace(event.Occurred), time.Now().Format(time.RFC3339))
	event.Service = normalizeDefault(strings.TrimSpace(event.Service), "router")
	event.EventType = normalizeDefault(strings.TrimSpace(event.EventType), strings.Trim(strings.Join([]string{event.Domain, event.Subtype}, "_"), "_"))
	event.Title = normalizeDefault(strings.TrimSpace(event.Title), domainTitle(event.Domain))
	event.Summary = normalizeDefault(strings.TrimSpace(event.Summary), event.Title)
	event.ErrorMessage = normalizeDefault(strings.TrimSpace(event.ErrorMessage), event.Summary)
	event.ImpactScope = normalizeDefault(strings.TrimSpace(event.ImpactScope), "unknown")
	event.ImpactSummary = normalizeDefault(strings.TrimSpace(event.ImpactSummary), "影响待排查")
	event.Title = truncateText(event.Title, 120)
	event.Summary = truncateText(event.Summary, 400)
	event.ErrorMessage = truncateText(event.ErrorMessage, 1200)
	event.ImpactSummary = truncateText(event.ImpactSummary, 800)
	if event.Tags == nil {
		event.Tags = map[string]string{}
	}
	return event
}

func buildErrorCardLine(event ErrorCardEvent) (string, error) {
	detailLines := []string{
		fmt.Sprintf("**事件类型**: `%s`", event.EventType),
		fmt.Sprintf("**错误分类**: `%s/%s`", event.Domain, event.Subtype),
		fmt.Sprintf("**错误码**: `%s`", event.ErrorCode),
		fmt.Sprintf("**业务结论**: `%s`", event.BizStatus),
		fmt.Sprintf("**影响范围**: `%s`", event.ImpactScope),
		fmt.Sprintf("**发生时间**: `%s`", event.Occurred),
	}
	if event.UserID != "" {
		detailLines = append(detailLines, fmt.Sprintf("**用户**: `%s`", event.UserID))
	}
	if event.GroupID != "" {
		detailLines = append(detailLines, fmt.Sprintf("**分组**: `%s`", event.GroupID))
	}
	if event.ChannelID != "" {
		detailLines = append(detailLines, fmt.Sprintf("**渠道**: `%s`", event.ChannelID))
	}
	if event.ModelName != "" {
		detailLines = append(detailLines, fmt.Sprintf("**模型**: `%s`", event.ModelName))
	}
	if event.Endpoint != "" {
		detailLines = append(detailLines, fmt.Sprintf("**接口**: `%s`", event.Endpoint))
	}
	if event.HTTPStatus != 0 {
		detailLines = append(detailLines, fmt.Sprintf("**HTTP 状态**: `%d`", event.HTTPStatus))
	}
	if event.UpstreamStatus != 0 {
		detailLines = append(detailLines, fmt.Sprintf("**上游状态**: `%d`", event.UpstreamStatus))
	}
	if event.ProviderCode != "" {
		detailLines = append(detailLines, fmt.Sprintf("**上游返回码**: `%s`", event.ProviderCode))
	}
	if event.ProviderStatus != "" {
		detailLines = append(detailLines, fmt.Sprintf("**上游状态信息**: `%s`", event.ProviderStatus))
	}
	if event.RetryCount > 0 {
		detailLines = append(detailLines, fmt.Sprintf("**重试次数**: `%d`", event.RetryCount))
	}
	if event.TraceID != "" {
		detailLines = append(detailLines, fmt.Sprintf("**Trace ID**: `%s`", event.TraceID))
	}
	if event.RequestID != "" {
		detailLines = append(detailLines, fmt.Sprintf("**Request ID**: `%s`", event.RequestID))
	}
	detailLines = appendTagLines(detailLines, event.Tags)

	elements := []any{
		map[string]any{
			"tag":     "markdown",
			"content": fmt.Sprintf("**问题**: %s\n**影响**: %s", event.ErrorMessage, event.ImpactSummary),
		},
		map[string]any{"tag": "hr"},
		map[string]any{
			"tag":     "markdown",
			"content": strings.Join(detailLines, "\n"),
		},
	}

	actions := buildCardActions(event)
	if len(actions) > 0 {
		elements = append(elements, map[string]any{
			"tag":     "action",
			"actions": actions,
		})
	}

	payload := map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"config": map[string]any{
				"wide_screen_mode": true,
				"enable_forward":   true,
			},
			"header": map[string]any{
				"template": severityTemplate(event.Severity),
				"title": map[string]any{
					"tag":     "plain_text",
					"content": fmt.Sprintf("【%s】%s", domainTitle(event.Domain), event.Title),
				},
			},
			"elements": elements,
		},
		"event": event,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func buildCardActions(event ErrorCardEvent) []any {
	actions := make([]any, 0, 4)
	if strings.TrimSpace(event.DetailURL) != "" {
		actions = append(actions, map[string]any{
			"tag":  "button",
			"type": "default",
			"text": map[string]any{"tag": "plain_text", "content": "查看详情"},
			"url":  event.DetailURL,
		})
	}
	if strings.TrimSpace(event.Runbook) != "" {
		actions = append(actions, map[string]any{
			"tag":  "button",
			"type": "default",
			"text": map[string]any{"tag": "plain_text", "content": "处理手册"},
			"url":  event.Runbook,
		})
	}
	for _, item := range event.Actions {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(item.Mode))
		if mode == "" {
			mode = "link"
		}
		button := map[string]any{
			"tag":  "button",
			"type": "default",
			"text": map[string]any{"tag": "plain_text", "content": label},
		}
		if strings.TrimSpace(item.URL) != "" {
			button["url"] = strings.TrimSpace(item.URL)
		}
		button["value"] = map[string]any{"id": strings.TrimSpace(item.ID), "mode": mode}
		actions = append(actions, button)
	}
	return actions
}

func writeErrorCardLine(line string) {
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	errorCardWriteMu.Lock()
	defer errorCardWriteMu.Unlock()
	var writer io.Writer
	switch {
	case errorWriter != nil:
		writer = errorWriter
	case routerErrorWriter != nil:
		writer = routerErrorWriter
	default:
		writer = nil
	}
	if writer == nil {
		return
	}
	_, _ = io.WriteString(writer, line)
}

func normalizeDomain(domain string) string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	switch normalized {
	case "channel", "payment", "group_billing", "router_internal":
		return normalized
	default:
		return "router_internal"
	}
}

func normalizeSeverity(severity string) string {
	normalized := strings.ToLower(strings.TrimSpace(severity))
	switch normalized {
	case "info", "warning", "error", "critical":
		return normalized
	default:
		return "error"
	}
}

func normalizeDefault(value string, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return strings.TrimSpace(defaultValue)
	}
	return strings.TrimSpace(value)
}

func severityTemplate(severity string) string {
	switch normalizeSeverity(severity) {
	case "critical":
		return "red"
	case "error":
		return "orange"
	case "warning":
		return "yellow"
	default:
		return "blue"
	}
}

func domainTitle(domain string) string {
	switch normalizeDomain(domain) {
	case "channel":
		return "渠道错误"
	case "payment":
		return "支付错误"
	case "group_billing":
		return "分组错误"
	default:
		return "系统错误"
	}
}

func IntString(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

func appendTagLines(lines []string, tags map[string]string) []string {
	if len(tags) == 0 {
		return lines
	}
	keys := make([]string, 0, len(tags))
	for key, value := range tags {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, strings.TrimSpace(key))
	}
	if len(keys) == 0 {
		return lines
	}
	sort.Strings(keys)
	seen := map[string]struct{}{}
	const maxVisibleTags = 12
	visible := 0
	for _, key := range keys {
		normalizedKey := strings.TrimSpace(key)
		if _, ok := seen[normalizedKey]; ok {
			continue
		}
		seen[normalizedKey] = struct{}{}
		value := strings.TrimSpace(tags[key])
		if value == "" {
			continue
		}
		if visible >= maxVisibleTags {
			lines = append(lines, fmt.Sprintf("**更多标签**: `%d`", len(keys)-visible))
			break
		}
		lines = append(lines, fmt.Sprintf("**%s**: `%s`", escapeCardCode(normalizedKey), escapeCardCode(truncateText(value, 200))))
		visible++
	}
	return lines
}

func escapeCardCode(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "`", "'")
}

func truncateText(value string, maxLen int) string {
	trimmed := strings.TrimSpace(value)
	if maxLen <= 0 || len(trimmed) <= maxLen {
		return trimmed
	}
	if maxLen <= 3 {
		return trimmed[:maxLen]
	}
	return trimmed[:maxLen-3] + "..."
}
