package monitor

import (
	"strings"
	"sync"
	"time"

	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
)

var store = make(map[string][]bool)
var metricSuccessChan = make(chan string, config.MetricSuccessChanSize)
var metricFailChan = make(chan string, config.MetricFailChanSize)
var metricStoreMu sync.Mutex
var metricRecoverTimers sync.Map

func consumeSuccess(channelId string) {
	metricStoreMu.Lock()
	defer metricStoreMu.Unlock()
	if len(store[channelId]) > config.MetricQueueSize {
		store[channelId] = store[channelId][1:]
	}
	store[channelId] = append(store[channelId], true)
}

func consumeFail(channelId string) (bool, float64) {
	metricStoreMu.Lock()
	defer metricStoreMu.Unlock()
	if len(store[channelId]) > config.MetricQueueSize {
		store[channelId] = store[channelId][1:]
	}
	store[channelId] = append(store[channelId], false)
	successCount := 0
	for _, success := range store[channelId] {
		if success {
			successCount++
		}
	}
	successRate := float64(successCount) / float64(len(store[channelId]))
	if len(store[channelId]) < config.MetricQueueSize {
		return false, successRate
	}
	if successRate < config.MetricSuccessRateThreshold {
		store[channelId] = make([]bool, 0)
		return true, successRate
	}
	return false, successRate
}

func metricSuccessConsumer() {
	for {
		select {
		case channelId := <-metricSuccessChan:
			consumeSuccess(channelId)
		}
	}
}

func metricFailConsumer() {
	for {
		select {
		case channelId := <-metricFailChan:
			disable, successRate := consumeFail(channelId)
			if disable {
				go MetricDisableChannelAndScheduleRecover(channelId, successRate)
			}
		}
	}
}

func init() {
	if config.EnableMetric {
		go metricSuccessConsumer()
		go metricFailConsumer()
	}
}

func Emit(channelId string, success bool) {
	if !config.EnableMetric {
		return
	}
	go func() {
		if success {
			metricSuccessChan <- channelId
		} else {
			metricFailChan <- channelId
		}
	}()
}

func MetricDisableChannelAndScheduleRecover(channelId string, successRate float64) {
	MetricDisableChannel(channelId, successRate)
	scheduleMetricChannelRecover(channelId)
}

func scheduleMetricChannelRecover(channelId string) {
	normalizedChannelID := strings.TrimSpace(channelId)
	if normalizedChannelID == "" {
		return
	}
	if !config.AutomaticEnableChannelEnabled {
		return
	}
	if config.MetricAutoRecoverAfterSeconds <= 0 {
		return
	}
	if _, loaded := metricRecoverTimers.LoadOrStore(normalizedChannelID, struct{}{}); loaded {
		return
	}
	time.AfterFunc(time.Duration(config.MetricAutoRecoverAfterSeconds)*time.Second, func() {
		metricRecoverTimers.Delete(normalizedChannelID)
		recoverMetricDisabledChannel(normalizedChannelID)
	})
}

func recoverMetricDisabledChannel(channelId string) {
	channel, err := model.GetChannelById(channelId)
	if err != nil {
		logger.SysError("failed to load channel for metric auto recover: " + err.Error())
		return
	}
	if channel.Status != model.ChannelStatusAutoDisabled {
		return
	}
	EnableChannel(channel.Id, channel.DisplayName())
}
