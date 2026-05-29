package channel

import (
	"testing"
	"time"

	"github.com/yeying-community/router/internal/admin/model"
)

func TestShouldDisableChannelForBillingEntitlementsDisablesPureBalanceChannel(t *testing.T) {
	items := []model.ChannelBillingSnapshotItem{
		{
			ResourceType:    model.ChannelBillingResourceTypeCredit,
			QuotaType:       "total",
			RemainingAmount: 0,
			Status:          model.ChannelBillingItemStatusDepleted,
		},
	}

	if !shouldDisableChannelForBillingEntitlements(collectedChannelBillingSnapshot{}, items, time.Now().Unix()) {
		t.Fatalf("pure balance channel with zero total remaining should be disabled")
	}
}

func TestShouldDisableChannelForBillingEntitlementsKeepsPackageChannelWithDailyQuota(t *testing.T) {
	now := time.Now().Unix()
	items := []model.ChannelBillingSnapshotItem{
		{
			ResourceType:    model.ChannelBillingResourceTypePlan,
			QuotaType:       "custom",
			RemainingAmount: 1,
			ExpiresAt:       now + 3600,
			Status:          model.ChannelBillingItemStatusActive,
		},
		{
			ResourceType:    model.ChannelBillingResourceTypeCredit,
			QuotaType:       "total",
			RemainingAmount: 0,
			Status:          model.ChannelBillingItemStatusDepleted,
		},
		{
			ResourceType:    model.ChannelBillingResourceTypeQuota,
			QuotaType:       "weekly",
			RemainingAmount: 0,
			Status:          model.ChannelBillingItemStatusDepleted,
		},
		{
			ResourceType:    model.ChannelBillingResourceTypeQuota,
			QuotaType:       "daily",
			RemainingAmount: 35.54,
			Status:          model.ChannelBillingItemStatusActive,
		},
	}

	if shouldDisableChannelForBillingEntitlements(collectedChannelBillingSnapshot{ShouldHardStop: true}, items, now) {
		t.Fatalf("package channel with usable daily quota should not be disabled by zero total credit")
	}
}

func TestShouldDisableChannelForBillingEntitlementsDisablesExhaustedPackageChannel(t *testing.T) {
	now := time.Now().Unix()
	items := []model.ChannelBillingSnapshotItem{
		{
			ResourceType:    model.ChannelBillingResourceTypePlan,
			QuotaType:       "custom",
			RemainingAmount: 1,
			ExpiresAt:       now - 1,
			Status:          model.ChannelBillingItemStatusExpired,
		},
		{
			ResourceType:    model.ChannelBillingResourceTypeQuota,
			QuotaType:       "daily",
			RemainingAmount: 0,
			Status:          model.ChannelBillingItemStatusDepleted,
		},
		{
			ResourceType:    model.ChannelBillingResourceTypeQuota,
			QuotaType:       "weekly",
			RemainingAmount: 0,
			Status:          model.ChannelBillingItemStatusDepleted,
		},
	}

	if !shouldDisableChannelForBillingEntitlements(collectedChannelBillingSnapshot{}, items, now) {
		t.Fatalf("package channel without usable plan or periodic quota should be disabled")
	}
}
