package billing

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
	channelsvc "github.com/yeying-community/router/internal/admin/service/channel"
)

const (
	channelBillingSchedulerTickSeconds          = 30
	channelBillingAutoRefreshMinIntervalSeconds = 60
)

var startChannelBillingAutoRefreshWorkerOnce sync.Once

func StartChannelBillingAutoRefreshWorker() {
	startChannelBillingAutoRefreshWorkerOnce.Do(func() {
		go runChannelBillingAutoRefreshWorker()
	})
}

func runChannelBillingAutoRefreshWorker() {
	logger.SysLog("[billing.channel] auto refresh worker started")
	ticker := time.NewTicker(channelBillingSchedulerTickSeconds * time.Second)
	defer ticker.Stop()

	for {
		if shouldRunChannelBillingAutoRefreshNow() {
			runChannelBillingAutoRefreshOnce()
		}
		<-ticker.C
	}
}

func shouldRunChannelBillingAutoRefreshNow() bool {
	if !config.ChannelBillingAutoRefreshEnabled {
		return false
	}
	now := helper.GetTimestamp()
	intervalSeconds := int64(normalizedChannelBillingAutoRefreshIntervalSeconds())
	if config.ChannelBillingAutoRefreshLastRunAt <= 0 {
		return true
	}
	return now-config.ChannelBillingAutoRefreshLastRunAt >= intervalSeconds
}

func runChannelBillingAutoRefreshOnce() {
	runAt := helper.GetTimestamp()
	_ = model.UpdateOption("ChannelBillingAutoRefreshLastRunAt", fmt.Sprintf("%d", runAt))

	channels, err := channelsvc.GetAllBasic(0, 0, "all", true)
	if err != nil {
		logger.SysWarnf("[billing.channel] list channels failed: %s", err.Error())
		return
	}

	submittedCount := 0
	reusedCount := 0
	failedCount := 0
	for _, channel := range channels {
		if channel == nil || strings.TrimSpace(channel.Id) == "" {
			continue
		}
		if channel.Status != model.ChannelStatusEnabled {
			continue
		}
		task, reused, err := model.CreateOrReuseAsyncTaskWithDB(model.DB, model.AsyncTask{
			Type:      model.AsyncTaskTypeChannelRefreshBilling,
			DedupeKey: fmt.Sprintf("%s:%s", model.AsyncTaskTypeChannelRefreshBilling, strings.TrimSpace(channel.Id)),
			ChannelId: strings.TrimSpace(channel.Id),
			Payload: marshalChannelBillingSchedulerPayload(map[string]any{
				"channel_id": strings.TrimSpace(channel.Id),
			}),
			CreatedBy: "",
			TraceID:   "",
		})
		if err != nil {
			failedCount++
			logger.SysWarnf("[billing.channel] enqueue refresh failed channel_id=%s err=%s", strings.TrimSpace(channel.Id), err.Error())
			continue
		}
		if reused {
			reusedCount++
			continue
		}
		if strings.TrimSpace(task.Id) != "" {
			submittedCount++
		}
	}

	logger.SysLogf("[billing.channel] auto refresh queued submitted=%d reused=%d failed=%d", submittedCount, reusedCount, failedCount)
}

func marshalChannelBillingSchedulerPayload(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func normalizedChannelBillingAutoRefreshIntervalSeconds() int {
	interval := config.ChannelBillingAutoRefreshIntervalSeconds
	if interval < channelBillingAutoRefreshMinIntervalSeconds {
		interval = channelBillingAutoRefreshMinIntervalSeconds
	}
	return interval
}
