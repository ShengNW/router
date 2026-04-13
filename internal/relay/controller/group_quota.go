package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
	"github.com/yeying-community/router/internal/relay/adaptor/openai"
	relaymodel "github.com/yeying-community/router/internal/relay/model"
)

const groupDailyQuotaExceededCode = "group_daily_quota_exceeded"

func formatGroupDailyQuotaExceededMessage(requested int64, snapshot model.GroupDailyQuotaSnapshot) string {
	requestedYYC := requested
	if requestedYYC < 0 {
		requestedYYC = 0
	}
	return fmt.Sprintf(
		"当前分组套餐每日额度不足：本次预估消耗 %d YYC，今日剩余 %d YYC（已用 %d，预占 %d，日上限 %d）",
		requestedYYC,
		snapshot.RemainingQuota,
		snapshot.ConsumedQuota,
		snapshot.ReservedQuota,
		snapshot.Limit,
	)
}

func formatPackageQuotaExceededMessage(requested int64, daily model.GroupDailyQuotaSnapshot, emergency model.UserPackageEmergencyQuotaSnapshot) string {
	requestedYYC := requested
	if requestedYYC < 0 {
		requestedYYC = 0
	}
	return fmt.Sprintf(
		"当前分组套餐额度不足：本次预估消耗 %d YYC，每日剩余 %d YYC（已用 %d，预占 %d，日上限 %d），应急剩余 %d YYC（已用 %d，预占 %d，应急上限 %d）",
		requestedYYC,
		daily.RemainingQuota,
		daily.ConsumedQuota,
		daily.ReservedQuota,
		daily.Limit,
		emergency.RemainingQuota,
		emergency.ConsumedQuota,
		emergency.ReservedQuota,
		emergency.Limit,
	)
}

func reservePackageQuota(ctx context.Context, groupID string, userID string, quota int64) (model.PackageQuotaReservation, *relaymodel.ErrorWithStatusCode) {
	reservation, allowed, err := model.ReservePackageQuota(groupID, userID, quota)
	if err != nil {
		emitGroupBillingErrorCard(ctx, "reserve_group_daily_quota_failed", "RESERVE_GROUP_DAILY_QUOTA_FAILED", "分组额度预占失败", err.Error(), strings.TrimSpace(groupID), strings.TrimSpace(userID), quota, "single_user", "当前请求无法预占分组额度，计费流程中断")
		return model.PackageQuotaReservation{}, openai.ErrorWrapper(err, "reserve_group_daily_quota_failed", http.StatusInternalServerError)
	}
	if !allowed {
		message := "当前分组套餐每日额度已达上限，请明日再试"
		dailySnapshot, dailyErr := model.GetGroupDailyQuotaSnapshot(groupID, userID, "")
		emergencySnapshot, emergencyErr := model.GetUserPackageEmergencyQuotaSnapshot(userID, "")
		if dailyErr != nil || emergencyErr != nil {
			logger.Warnf(ctx, "package quota denied group=%s user=%s requested=%d daily_snapshot_err=%v emergency_snapshot_err=%v", strings.TrimSpace(groupID), strings.TrimSpace(userID), quota, dailyErr, emergencyErr)
		} else {
			logger.Warnf(
				ctx,
				"package quota denied group=%s user=%s biz_date=%s biz_month=%s requested=%d daily_limit=%d daily_consumed=%d daily_reserved=%d daily_remaining=%d emergency_limit=%d emergency_consumed=%d emergency_reserved=%d emergency_remaining=%d",
				dailySnapshot.GroupID,
				dailySnapshot.UserID,
				dailySnapshot.BizDate,
				emergencySnapshot.BizMonth,
				quota,
				dailySnapshot.Limit,
				dailySnapshot.ConsumedQuota,
				dailySnapshot.ReservedQuota,
				dailySnapshot.RemainingQuota,
				emergencySnapshot.Limit,
				emergencySnapshot.ConsumedQuota,
				emergencySnapshot.ReservedQuota,
				emergencySnapshot.RemainingQuota,
			)
			message = formatPackageQuotaExceededMessage(quota, dailySnapshot, emergencySnapshot)
		}
		emitGroupBillingErrorCard(ctx, groupDailyQuotaExceededCode, strings.ToUpper(groupDailyQuotaExceededCode), "分组额度不足", message, strings.TrimSpace(groupID), strings.TrimSpace(userID), quota, "single_user", "当前请求被分组/套餐额度限制拦截")
		return model.PackageQuotaReservation{}, openai.ErrorWrapper(errors.New(message), groupDailyQuotaExceededCode, http.StatusForbidden)
	}
	return reservation, nil
}

func releasePackageQuotaReservation(ctx context.Context, reservation model.PackageQuotaReservation) {
	if !reservation.Active() {
		return
	}
	if err := model.ReleasePackageQuotaReservation(reservation); err != nil {
		logger.Error(ctx, "release package quota reservation failed: "+err.Error())
		emitGroupBillingErrorCard(ctx, "release_package_quota_reservation_failed", "RELEASE_PACKAGE_QUOTA_RESERVATION_FAILED", "分组额度释放失败", err.Error(), strings.TrimSpace(reservation.GroupDaily.GroupID), strings.TrimSpace(reservation.GroupDaily.UserID), reservation.GroupDaily.ReservedQuota+reservation.PackageEmergency.ReservedQuota, "single_user", "预占额度释放失败，可能影响后续可用额度计算")
	}
}

func settlePackageQuotaReservation(ctx context.Context, reservation model.PackageQuotaReservation, consumedQuota int64) (int64, int64) {
	if !reservation.Active() {
		return 0, 0
	}
	dailyConsumed, emergencyConsumed, err := model.SettlePackageQuotaReservation(reservation, consumedQuota)
	if err != nil {
		logger.Error(ctx, "settle package quota reservation failed: "+err.Error())
		emitGroupBillingErrorCard(ctx, "settle_package_quota_reservation_failed", "SETTLE_PACKAGE_QUOTA_RESERVATION_FAILED", "分组额度结算失败", err.Error(), strings.TrimSpace(reservation.GroupDaily.GroupID), strings.TrimSpace(reservation.GroupDaily.UserID), consumedQuota, "single_user", "本次调用已完成但分组额度结算失败，账务可能不一致")
		return 0, 0
	}
	return dailyConsumed, emergencyConsumed
}

func IsGroupDailyQuotaExceededError(err *relaymodel.ErrorWithStatusCode) bool {
	if err == nil {
		return false
	}
	code := strings.TrimSpace(fmt.Sprint(err.Code))
	return code == groupDailyQuotaExceededCode
}

func emitGroupBillingErrorCard(ctx context.Context, subtype string, errorCode string, title string, message string, groupID string, userID string, quota int64, impactScope string, impactSummary string) {
	tags := map[string]string{}
	if quota != 0 {
		tags["quota"] = strconv.FormatInt(quota, 10)
	}
	logger.EmitFeishuCardError(ctx, logger.ErrorCardEvent{
		EventType:     "group_billing_error",
		Domain:        "group_billing",
		Subtype:       strings.TrimSpace(subtype),
		Severity:      "error",
		Title:         strings.TrimSpace(title),
		Summary:       strings.TrimSpace(message),
		BizStatus:     "failed",
		ErrorCode:     strings.TrimSpace(errorCode),
		ErrorMessage:  strings.TrimSpace(message),
		ImpactScope:   strings.TrimSpace(impactScope),
		ImpactSummary: strings.TrimSpace(impactSummary),
		UserID:        strings.TrimSpace(userID),
		GroupID:       strings.TrimSpace(groupID),
		Tags:          tags,
	})
}
