package model

import (
	"fmt"

	"gorm.io/gorm"
)

func runDropChannelTypeColumnMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if !tx.Migrator().HasTable(&Channel{}) {
		return nil
	}
	if tx.Migrator().HasColumn(&Channel{}, "type") {
		if err := tx.Migrator().DropColumn(&Channel{}, "type"); err != nil {
			return err
		}
	}
	return nil
}
