package common

import (
	"fmt"
	"strings"

	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/yeying-community/router/common/config"
)

func LogQuota(quota int64) string {
	if config.DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/config.QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}

// IsValidEthAddress performs a basic checksum/length check
func IsValidEthAddress(addr string) bool {
	if addr == "" {
		return false
	}
	return gethCommon.IsHexAddress(strings.ToLower(addr))
}
