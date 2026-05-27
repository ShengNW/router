package monitor

import (
	"strings"
	"sync"
	"time"

	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
)

var store = make(map[string][]bool)
var metricSuccessChan = make(chan string, config.MetricSuccessChanSize)
var metricFailChan = make(chan string, config.MetricFailChanSize)
var metricStoreMu sync.Mutex
var metricRecoverTimers sync.Map
var metricConsumersOnce sync.Once

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

func StartMetricMonitor() {
	if !config.EnableMetric {
		return
	}
	metricConsumersOnce.Do(func() {
		go metricSuccessConsumer()
		go metricFailConsumer()
	})
	resumeMetricChannelRecoveries()
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
	normalizedChannelID := strings.TrimSpace(channelId)
	if normalizedChannelID == "" {
		return
	}
	MetricDisableChannel(channelId, successRate)
	recoverAfter := helper.GetTimestamp() + int64(config.MetricAutoRecoverAfterSeconds)
	if err := model.RecordChannelCircuitBreakerOpen(normalizedChannelID, "low_success_rate", successRate, recoverAfter); err != nil {
		logger.SysError("failed to record metric circuit breaker state: " + err.Error())
	}
	scheduleMetricChannelRecoverAt(normalizedChannelID, recoverAfter)
}

func scheduleMetricChannelRecoverAt(channelId string, recoverAfter int64) {
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
	delaySeconds := recoverAfter - helper.GetTimestamp()
	if delaySeconds < 0 {
		delaySeconds = 0
	}
	time.AfterFunc(time.Duration(delaySeconds)*time.Second, func() {
		metricRecoverTimers.Delete(normalizedChannelID)
		recoverMetricDisabledChannel(normalizedChannelID)
	})
}

func resumeMetricChannelRecoveries() {
	if !config.AutomaticEnableChannelEnabled {
		return
	}
	rows, err := model.ListOpenChannelCircuitBreakerStates()
	if err != nil {
		logger.SysError("failed to list metric circuit breaker states: " + err.Error())
		return
	}
	for _, row := range rows {
		scheduleMetricChannelRecoverAt(row.ChannelId, row.RecoverAfter)
	}
}

func recoverMetricDisabledChannel(channelId string) {
	state, err := model.GetChannelCircuitBreakerState(channelId)
	if err != nil {
		logger.SysError("failed to load metric circuit breaker state: " + err.Error())
		return
	}
	if state.State != model.ChannelCircuitBreakerStateOpen {
		return
	}
	channel, err := model.GetChannelById(channelId)
	if err != nil {
		logger.SysError("failed to load channel for metric auto recover: " + err.Error())
		return
	}
	if channel.Status != model.ChannelStatusAutoDisabled {
		return
	}
	RecoverMetricDisabledChannel(channel.Id, channel.DisplayName())
	if err := model.RecordChannelCircuitBreakerRecovered(channel.Id); err != nil {
		logger.SysError("failed to record metric circuit breaker recovery: " + err.Error())
	}
}
