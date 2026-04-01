package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/internal/admin/model"
)

const (
	fxProviderFrankfurter       = "frankfurter"
	fxFrankfurterLatestEndpoint = "https://api.frankfurter.app/latest"
	fxSyncTimeout               = 15 * time.Second
)

type FXSyncUpdatedCurrency struct {
	Code          string  `json:"code"`
	USDRate       float64 `json:"usd_rate"`
	OldYYCPerUnit float64 `json:"old_yyc_per_unit"`
	NewYYCPerUnit float64 `json:"new_yyc_per_unit"`
}

type FXSyncSkippedCurrency struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type FXSyncResult struct {
	Provider     string                  `json:"provider"`
	Base         string                  `json:"base"`
	Date         string                  `json:"date"`
	UpdatedCount int                     `json:"updated_count"`
	SkippedCount int                     `json:"skipped_count"`
	Updated      []FXSyncUpdatedCurrency `json:"updated"`
	Skipped      []FXSyncSkippedCurrency `json:"skipped"`
}

type frankfurterLatestResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

func SyncBillingCurrenciesFromFX(ctx context.Context) (FXSyncResult, error) {
	result := FXSyncResult{
		Provider: fxProviderFrankfurter,
		Base:     model.BillingCurrencyCodeUSD,
		Updated:  make([]FXSyncUpdatedCurrency, 0),
		Skipped:  make([]FXSyncSkippedCurrency, 0),
	}
	if model.DB == nil {
		return result, fmt.Errorf("database handle is nil")
	}

	usdYYCPerUnit, err := model.GetBillingCurrencyYYCPerUnit(model.BillingCurrencyCodeUSD)
	if err != nil {
		return result, fmt.Errorf("failed to load USD yyc rate: %w", err)
	}

	rows, err := model.ListBillingCurrencies()
	if err != nil {
		return result, err
	}
	rowByCode := make(map[string]model.BillingCurrency, len(rows))
	candidateSet := make(map[string]struct{})
	appendSkipped := func(code string, reason string) {
		result.Skipped = append(result.Skipped, FXSyncSkippedCurrency{
			Code:   code,
			Reason: reason,
		})
	}

	for _, row := range rows {
		code := strings.ToUpper(strings.TrimSpace(row.Code))
		if code == "" {
			continue
		}
		rowByCode[code] = row
		if code == model.BillingCurrencyCodeUSD {
			appendSkipped(code, "base_currency")
			continue
		}
		source := strings.ToLower(strings.TrimSpace(row.Source))
		if source == model.BillingCurrencySourceManual {
			appendSkipped(code, "manual_locked")
			continue
		}
		if source != model.BillingCurrencySourceSystemDefault && source != model.BillingCurrencySourceFXAuto {
			appendSkipped(code, "source_not_auto_managed")
			continue
		}
		candidateSet[code] = struct{}{}
	}

	if len(candidateSet) == 0 {
		result.UpdatedCount = 0
		result.SkippedCount = len(result.Skipped)
		return result, nil
	}

	candidateCodes := make([]string, 0, len(candidateSet))
	for code := range candidateSet {
		candidateCodes = append(candidateCodes, code)
	}
	sort.Strings(candidateCodes)

	ratePayload, err := fetchFrankfurterLatestRates(ctx, result.Base, candidateCodes)
	if err != nil {
		return result, err
	}
	if strings.TrimSpace(ratePayload.Date) != "" {
		result.Date = ratePayload.Date
	}
	now := helper.GetTimestamp()
	for _, code := range candidateCodes {
		rate, ok := ratePayload.Rates[code]
		if !ok || rate <= 0 {
			appendSkipped(code, "rate_not_found")
			continue
		}
		nextYYCPerUnit := usdYYCPerUnit / rate
		if math.IsNaN(nextYYCPerUnit) || math.IsInf(nextYYCPerUnit, 0) || nextYYCPerUnit <= 0 {
			appendSkipped(code, "invalid_rate_value")
			continue
		}
		tx := model.DB.Model(&model.BillingCurrency{}).
			Where("code = ? AND lower(trim(source)) IN ?", code, []string{
				model.BillingCurrencySourceSystemDefault,
				model.BillingCurrencySourceFXAuto,
			}).
			Updates(map[string]any{
				"yyc_per_unit": nextYYCPerUnit,
				"source":       model.BillingCurrencySourceFXAuto,
				"updated_at":   now,
			})
		if tx.Error != nil {
			appendSkipped(code, "update_failed")
			continue
		}
		if tx.RowsAffected == 0 {
			appendSkipped(code, "source_changed_or_missing")
			continue
		}

		oldYYCPerUnit := 0.0
		if oldRow, ok := rowByCode[code]; ok {
			oldYYCPerUnit = oldRow.YYCPerUnit
		}
		result.Updated = append(result.Updated, FXSyncUpdatedCurrency{
			Code:          code,
			USDRate:       rate,
			OldYYCPerUnit: oldYYCPerUnit,
			NewYYCPerUnit: nextYYCPerUnit,
		})
	}

	result.UpdatedCount = len(result.Updated)
	result.SkippedCount = len(result.Skipped)
	if err := model.SyncBillingCurrencyCatalogWithDB(model.DB); err != nil {
		return result, err
	}
	return result, nil
}

func fetchFrankfurterLatestRates(ctx context.Context, base string, targetCodes []string) (frankfurterLatestResponse, error) {
	normalizedBase := strings.ToUpper(strings.TrimSpace(base))
	if normalizedBase == "" {
		normalizedBase = model.BillingCurrencyCodeUSD
	}

	normalizedTargets := make([]string, 0, len(targetCodes))
	seen := make(map[string]struct{}, len(targetCodes))
	for _, rawCode := range targetCodes {
		code := strings.ToUpper(strings.TrimSpace(rawCode))
		if code == "" || code == normalizedBase {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		normalizedTargets = append(normalizedTargets, code)
	}
	sort.Strings(normalizedTargets)
	if len(normalizedTargets) == 0 {
		return frankfurterLatestResponse{}, fmt.Errorf("no target currency for FX sync")
	}

	query := url.Values{}
	query.Set("from", normalizedBase)
	query.Set("to", strings.Join(normalizedTargets, ","))
	requestURL := fxFrankfurterLatestEndpoint + "?" + query.Encode()

	if ctx == nil {
		ctx = context.Background()
	}
	requestCtx, cancel := context.WithTimeout(ctx, fxSyncTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, requestURL, nil)
	if err != nil {
		return frankfurterLatestResponse{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return frankfurterLatestResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return frankfurterLatestResponse{}, fmt.Errorf("fx upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	payload := frankfurterLatestResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return frankfurterLatestResponse{}, err
	}
	if len(payload.Rates) == 0 {
		return frankfurterLatestResponse{}, fmt.Errorf("fx upstream returned empty rates")
	}
	normalizedRates := make(map[string]float64, len(payload.Rates))
	for code, rate := range payload.Rates {
		normalizedRates[strings.ToUpper(strings.TrimSpace(code))] = rate
	}
	payload.Rates = normalizedRates
	payload.Base = strings.ToUpper(strings.TrimSpace(payload.Base))
	return payload, nil
}
