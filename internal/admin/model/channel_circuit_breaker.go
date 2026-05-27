package model

import (
	"fmt"
	"strings"

	"github.com/yeying-community/router/common/helper"
	"gorm.io/gorm"
)

const (
	ChannelCircuitBreakerStateOpen      = "open"
	ChannelCircuitBreakerStateRecovered = "recovered"
	ChannelCircuitBreakerStateCanceled  = "canceled"
)

type ChannelCircuitBreakerState struct {
	ChannelId    string  `json:"channel_id" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	State        string  `json:"state" gorm:"type:varchar(32);not null;default:'open';index"`
	Reason       string  `json:"reason" gorm:"type:text"`
	SuccessRate  float64 `json:"success_rate" gorm:"type:double precision;default:0"`
	DisabledAt   int64   `json:"disabled_at" gorm:"bigint;index"`
	RecoverAfter int64   `json:"recover_after" gorm:"bigint;index"`
	RecoveredAt  int64   `json:"recovered_at" gorm:"bigint;index"`
	UpdatedAt    int64   `json:"updated_at" gorm:"bigint;index"`
}

func (ChannelCircuitBreakerState) TableName() string {
	return "channel_circuit_breaker_states"
}

func RecordChannelCircuitBreakerOpen(channelID string, reason string, successRate float64, recoverAfter int64) error {
	return recordChannelCircuitBreakerOpenWithDB(DB, channelID, reason, successRate, recoverAfter)
}

func RecordChannelCircuitBreakerRecovered(channelID string) error {
	return updateChannelCircuitBreakerStateWithDB(DB, channelID, ChannelCircuitBreakerStateRecovered, "")
}

func RecordChannelCircuitBreakerCanceled(channelID string, reason string) error {
	return updateChannelCircuitBreakerStateWithDB(DB, channelID, ChannelCircuitBreakerStateCanceled, reason)
}

func GetChannelCircuitBreakerState(channelID string) (ChannelCircuitBreakerState, error) {
	return getChannelCircuitBreakerStateWithDB(DB, channelID)
}

func ListOpenChannelCircuitBreakerStates() ([]ChannelCircuitBreakerState, error) {
	return listOpenChannelCircuitBreakerStatesWithDB(DB)
}

func recordChannelCircuitBreakerOpenWithDB(db *gorm.DB, channelID string, reason string, successRate float64, recoverAfter int64) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return nil
	}
	now := helper.GetTimestamp()
	row := ChannelCircuitBreakerState{
		ChannelId:    normalizedChannelID,
		State:        ChannelCircuitBreakerStateOpen,
		Reason:       strings.TrimSpace(reason),
		SuccessRate:  successRate,
		DisabledAt:   now,
		RecoverAfter: recoverAfter,
		RecoveredAt:  0,
		UpdatedAt:    now,
	}
	return db.Save(&row).Error
}

func updateChannelCircuitBreakerStateWithDB(db *gorm.DB, channelID string, state string, reason string) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	normalizedState := strings.TrimSpace(state)
	if normalizedChannelID == "" || normalizedState == "" {
		return nil
	}
	now := helper.GetTimestamp()
	updates := map[string]any{
		"state":        normalizedState,
		"recovered_at": now,
		"updated_at":   now,
	}
	if normalizedReason := strings.TrimSpace(reason); normalizedReason != "" {
		updates["reason"] = normalizedReason
	}
	return db.Model(&ChannelCircuitBreakerState{}).
		Where("channel_id = ? AND state = ?", normalizedChannelID, ChannelCircuitBreakerStateOpen).
		Updates(updates).Error
}

func getChannelCircuitBreakerStateWithDB(db *gorm.DB, channelID string) (ChannelCircuitBreakerState, error) {
	if db == nil {
		return ChannelCircuitBreakerState{}, fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return ChannelCircuitBreakerState{}, fmt.Errorf("channel id is empty")
	}
	row := ChannelCircuitBreakerState{}
	err := db.First(&row, "channel_id = ?", normalizedChannelID).Error
	return row, err
}

func listOpenChannelCircuitBreakerStatesWithDB(db *gorm.DB) ([]ChannelCircuitBreakerState, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	rows := make([]ChannelCircuitBreakerState, 0)
	err := db.Where("state = ?", ChannelCircuitBreakerStateOpen).Find(&rows).Error
	return rows, err
}
