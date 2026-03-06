package model

import (
	"fmt"

	"gorm.io/gorm"
)

func runRemoveOpenAICompatibleProtocolMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}

	if tx.Migrator().HasTable(&Channel{}) {
		if err := tx.Model(&Channel{}).
			Where("protocol = ?", "openai-compatible").
			Update("protocol", "openai").Error; err != nil {
			return err
		}
	}

	if tx.Migrator().HasTable(&ChannelProtocolCatalog{}) {
		if err := tx.Where("name = ?", "openai-compatible").
			Delete(&ChannelProtocolCatalog{}).Error; err != nil {
			return err
		}
	}

	return nil
}
