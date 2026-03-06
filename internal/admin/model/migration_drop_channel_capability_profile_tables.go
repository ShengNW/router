package model

import "gorm.io/gorm"

func dropChannelCapabilityProfileTablesWithDB(tx *gorm.DB) error {
	if tx == nil {
		return nil
	}
	if err := tx.Migrator().DropTable("channel_capability_profiles"); err != nil {
		return err
	}
	return tx.Migrator().DropTable("client_profiles")
}
