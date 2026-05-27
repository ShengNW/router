package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	ProviderModelsTableName = "provider_models"

	ProviderModelStatusActive     = "active"
	ProviderModelStatusDeprecated = "deprecated"
)

type ProviderModel struct {
	Provider           string  `json:"provider" gorm:"primaryKey;type:varchar(64)"`
	Model              string  `json:"model" gorm:"primaryKey;type:varchar(255)"`
	Tags               string  `json:"tags" gorm:"type:text;default:''"`
	Status             string  `json:"status" gorm:"type:varchar(32);not null;default:'active'"`
	Description        string  `json:"description" gorm:"type:text;default:''"`
	IsDeleted          bool    `json:"is_deleted" gorm:"not null;default:false"`
	SupportedEndpoints string  `json:"supported_endpoints" gorm:"type:text;default:''"`
	InputPrice         float64 `json:"input_price" gorm:"type:double precision;default:0"`
	OutputPrice        float64 `json:"output_price" gorm:"type:double precision;default:0"`
	PriceUnit          string  `json:"price_unit" gorm:"type:varchar(64);default:'per_1k_tokens'"`
	Currency           string  `json:"currency" gorm:"type:varchar(16);default:'USD'"`
	Source             string  `json:"source" gorm:"type:varchar(32);default:'manual'"`
	UpdatedAt          int64   `json:"updated_at" gorm:"bigint"`
}

func (ProviderModel) TableName() string {
	return ProviderModelsTableName
}

func ListActiveProviderModelsWithDB(db *gorm.DB, provider string) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	normalizedProvider := NormalizeGroupModelProviderValue(provider)
	if normalizedProvider == "" {
		return []string{}, nil
	}
	rows := make([]string, 0)
	if err := db.Model(&ProviderModel{}).
		Where("provider = ? AND is_deleted = ?", normalizedProvider, false).
		Order("model asc").
		Pluck("model", &rows).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		modelName := strings.TrimSpace(row)
		if modelName == "" {
			continue
		}
		result = append(result, modelName)
	}
	return result, nil
}
