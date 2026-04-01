package model

import "gorm.io/gorm"

func dropLegacyGroupQuotaColumnsWithDB(tx *gorm.DB) error {
	if tx == nil {
		return nil
	}
	if !tx.Migrator().HasTable((&GroupCatalog{}).TableName()) {
		return nil
	}
	if tx.Migrator().HasColumn((&GroupCatalog{}).TableName(), "daily_quota_limit") {
		if err := tx.Migrator().DropColumn((&GroupCatalog{}).TableName(), "daily_quota_limit"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn((&GroupCatalog{}).TableName(), "quota_reset_timezone") {
		if err := tx.Migrator().DropColumn((&GroupCatalog{}).TableName(), "quota_reset_timezone"); err != nil {
			return err
		}
	}
	return nil
}
