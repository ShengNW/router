package model

import (
	"fmt"

	relaychannel "github.com/yeying-community/router/internal/relay/channel"
	"gorm.io/gorm"
)

func runChannelProtocolMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if !tx.Migrator().HasTable(&Channel{}) {
		return nil
	}
	if err := tx.AutoMigrate(&Channel{}); err != nil {
		return err
	}

	hasTypeColumn := tx.Migrator().HasColumn(&Channel{}, "type")
	type row struct {
		Id       string `gorm:"column:id"`
		Protocol string `gorm:"column:protocol"`
		Type     int    `gorm:"column:type"`
	}
	rows := make([]row, 0)
	if hasTypeColumn {
		if err := tx.Model(&Channel{}).Select("id", "protocol", "type").Find(&rows).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Model(&Channel{}).Select("id", "protocol").Find(&rows).Error; err != nil {
			return err
		}
	}

	for _, item := range rows {
		protocol := relaychannel.NormalizeProtocolName(item.Protocol)
		if protocol == "" && hasTypeColumn {
			protocol = relaychannel.ProtocolByType(item.Type)
		}
		if protocol == "" {
			protocol = "openai"
		}
		updates := map[string]any{
			"protocol": protocol,
		}
		if err := tx.Model(&Channel{}).Where("id = ?", item.Id).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}
