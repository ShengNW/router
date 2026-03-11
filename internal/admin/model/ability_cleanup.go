package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func CleanupDanglingAbilityChannels() (int64, error) {
	return cleanupDanglingAbilityChannelsWithDB(DB)
}

func cleanupDanglingAbilityChannelsWithDB(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database handle is nil")
	}

	channelIDSubQuery := db.Model(&Channel{}).Select("id")
	result := db.
		Where("channel_id <> ''").
		Where("channel_id NOT IN (?)", channelIDSubQuery).
		Delete(&Ability{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func filterEnabledGroupAbilityRows(rows []Ability) []Ability {
	if len(rows) == 0 {
		return []Ability{}
	}
	filtered := make([]Ability, 0, len(rows))
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		if strings.TrimSpace(row.ChannelId) == "" {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}
