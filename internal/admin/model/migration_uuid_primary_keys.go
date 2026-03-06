package model

import (
	"fmt"

	"gorm.io/gorm"
)

func runUUIDPrimaryKeyDestructiveMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}

	// Full destructive migration requested by operator:
	// drop and rebuild core business tables with UUID primary keys.
	dropTargets := []interface{}{
		"abilities", // legacy table name before rename
		&Ability{},
		&Log{},
		&Token{},
		&Redemption{},
		&Channel{},
		&User{},
	}
	for _, target := range dropTargets {
		if tx.Migrator().HasTable(target) {
			if err := tx.Migrator().DropTable(target); err != nil {
				return err
			}
		}
	}

	return tx.AutoMigrate(
		&User{},
		&Channel{},
		&Token{},
		&Redemption{},
		&Ability{},
		&Log{},
	)
}

func runMainBaselineMigrationWithDB(tx *gorm.DB) error {
	if err := runUUIDPrimaryKeyDestructiveMigrationWithDB(tx); err != nil {
		return err
	}
	if err := runModelProviderMigrationsWithDB(tx); err != nil {
		return err
	}
	if err := runChannelProtocolCatalogMigrationsWithDB(tx); err != nil {
		return err
	}
	if err := runGroupCatalogMigrationsWithDB(tx); err != nil {
		return err
	}
	return runChannelTestModelMigrationsWithDB(tx)
}

func runLogUUIDPrimaryKeyDestructiveMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if tx.Migrator().HasTable(&Log{}) {
		if err := tx.Migrator().DropTable(&Log{}); err != nil {
			return err
		}
	}
	return tx.AutoMigrate(&Log{})
}
