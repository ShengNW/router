package model

import (
	"strings"

	commonutils "github.com/yeying-community/router/common/utils"
	"gorm.io/gorm"
)

func canonicalizeModelNameForProvider(provider string, modelName string) string {
	normalizedProvider := commonutils.NormalizeProvider(provider)
	if normalizedProvider == "" {
		normalizedProvider = strings.TrimSpace(strings.ToLower(provider))
	}
	name := strings.TrimSpace(modelName)
	if name == "" {
		return ""
	}
	if strings.Contains(name, "/") {
		parts := strings.SplitN(name, "/", 2)
		if len(parts) == 2 {
			prefix := commonutils.NormalizeProvider(parts[0])
			if prefix == "" || prefix == "unknown" {
				prefix = strings.TrimSpace(strings.ToLower(parts[0]))
			}
			if prefix == normalizedProvider {
				trimmed := strings.TrimSpace(parts[1])
				if trimmed != "" {
					name = trimmed
				}
			}
		}
	}
	lower := strings.ToLower(name)
	if normalizedProvider == "meta" && strings.HasPrefix(lower, "meta-") {
		trimmed := strings.TrimSpace(name[len("meta-"):])
		if trimmed != "" {
			name = trimmed
		}
	}
	return name
}

func LoadProviderModelDetailsMap(db *gorm.DB) (map[string][]ProviderModelDetail, error) {
	return LoadProviderModelDetailsMapForProviders(db, nil)
}

func LoadProviderModelDetailsMapForProviders(db *gorm.DB, providers []string) (map[string][]ProviderModelDetail, error) {
	rows := make([]ProviderModel, 0)
	query := db.Order("provider asc, model asc")
	if len(providers) > 0 {
		query = query.Where("provider IN ?", providers)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string][]ProviderModelDetail, 0)
	for _, row := range rows {
		provider := commonutils.NormalizeProvider(row.Provider)
		if provider == "" {
			provider = strings.TrimSpace(strings.ToLower(row.Provider))
		}
		if provider == "" {
			continue
		}
		modelName := canonicalizeModelNameForProvider(provider, row.Model)
		if modelName == "" {
			continue
		}
		result[provider] = append(result[provider], ProviderModelDetail{
			Model:       modelName,
			Type:        strings.TrimSpace(strings.ToLower(row.Type)),
			InputPrice:  row.InputPrice,
			OutputPrice: row.OutputPrice,
			PriceUnit:   strings.TrimSpace(strings.ToLower(row.PriceUnit)),
			Currency:    strings.TrimSpace(strings.ToUpper(row.Currency)),
			Source:      strings.TrimSpace(strings.ToLower(row.Source)),
			UpdatedAt:   row.UpdatedAt,
		})
	}
	for provider, details := range result {
		result[provider] = NormalizeProviderModelDetails(details)
	}
	return result, nil
}

func BuildProviderModelRows(provider string, details []ProviderModelDetail, now int64) []ProviderModel {
	normalizedProvider := commonutils.NormalizeProvider(provider)
	if normalizedProvider == "" {
		normalizedProvider = strings.TrimSpace(strings.ToLower(provider))
	}
	if normalizedProvider == "" {
		return nil
	}
	detailInput := make([]ProviderModelDetail, 0, len(details))
	for _, detail := range details {
		detail.Model = canonicalizeModelNameForProvider(normalizedProvider, detail.Model)
		if strings.TrimSpace(detail.Model) == "" {
			continue
		}
		detailInput = append(detailInput, detail)
	}
	normalizedDetails := NormalizeProviderModelDetails(detailInput)
	rows := make([]ProviderModel, 0, len(normalizedDetails))
	for _, detail := range normalizedDetails {
		updatedAt := detail.UpdatedAt
		if updatedAt == 0 {
			updatedAt = now
		}
		rows = append(rows, ProviderModel{
			Provider:    normalizedProvider,
			Model:       detail.Model,
			Type:        detail.Type,
			InputPrice:  detail.InputPrice,
			OutputPrice: detail.OutputPrice,
			PriceUnit:   detail.PriceUnit,
			Currency:    detail.Currency,
			Source:      detail.Source,
			UpdatedAt:   updatedAt,
		})
	}
	return rows
}
