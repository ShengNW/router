package model

import "gorm.io/gorm"

func renameModelProvidersTableToProviders(tx *gorm.DB) error {
	if tx == nil {
		return nil
	}
	if tx.Migrator().HasTable("providers") {
		return nil
	}
	if !tx.Migrator().HasTable("model_providers") {
		return nil
	}
	return tx.Migrator().RenameTable("model_providers", "providers")
}
