package model

import (
	"fmt"

	"gorm.io/gorm"
)

func runChannelCapabilityProfilesMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if err := tx.AutoMigrate(
		&ClientProfile{},
		&ChannelCapabilityProfile{},
	); err != nil {
		return err
	}
	return syncClientProfilesWithDB(tx)
}
