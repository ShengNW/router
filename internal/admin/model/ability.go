package model

import (
	"context"
	"strings"
)

type Ability struct {
	Group         string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model         string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId     string `json:"channel_id" gorm:"type:varchar(64);primaryKey;autoIncrement:false;index"`
	UpstreamModel string `json:"upstream_model" gorm:"type:varchar(255);default:'';index"`
	Enabled       bool   `json:"enabled"`
	Priority      *int64 `json:"priority" gorm:"bigint;default:0;index"`
}

const (
	AbilityTableName = "group_model_channels"
)

func (Ability) TableName() string {
	return AbilityTableName
}

func GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	return mustAbilityRepo().GetRandomSatisfiedChannel(group, model, ignoreFirstPriority)
}

func ListSatisfiedChannels(group string, model string) ([]*Channel, error) {
	return mustAbilityRepo().ListSatisfiedChannels(group, model)
}

func (channel *Channel) AddAbilities() error {
	return mustAbilityRepo().AddAbilities(channel)
}

func (channel *Channel) DeleteAbilities() error {
	return mustAbilityRepo().DeleteAbilities(channel)
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	return mustAbilityRepo().UpdateAbilities(channel)
}

func UpdateAbilityStatus(channelId string, status bool) error {
	return mustAbilityRepo().UpdateAbilityStatus(channelId, status)
}

// GetTopChannelByModel returns the highest-priority enabled channel for a given group+model.
// Order: priority desc, then channel_id asc (stable for UI usage).
func GetTopChannelByModel(group string, model string) (*Channel, error) {
	return mustAbilityRepo().GetTopChannelByModel(group, model)
}

func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	return mustAbilityRepo().GetGroupModels(ctx, group)
}

func NormalizeAbilityUpstreamModel(modelName string, upstreamModel string) string {
	upstream := strings.TrimSpace(upstreamModel)
	if upstream != "" {
		return upstream
	}
	return strings.TrimSpace(modelName)
}

func normalizeAbilityRowsPreserveOrder(rows []Ability) []Ability {
	if len(rows) == 0 {
		return []Ability{}
	}
	result := make([]Ability, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		normalized := row
		normalized.Group = strings.TrimSpace(normalized.Group)
		normalized.Model = strings.TrimSpace(normalized.Model)
		normalized.ChannelId = strings.TrimSpace(normalized.ChannelId)
		normalized.UpstreamModel = NormalizeAbilityUpstreamModel(normalized.Model, normalized.UpstreamModel)
		if normalized.Group == "" || normalized.Model == "" || normalized.ChannelId == "" {
			continue
		}
		key := normalized.Group + "::" + normalized.Model + "::" + normalized.ChannelId
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
