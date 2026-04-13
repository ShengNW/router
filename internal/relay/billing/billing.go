package billing

import (
	"context"
	"strconv"
	"strings"

	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
)

func ReturnPreConsumedQuota(ctx context.Context, preConsumedQuota int64, tokenId string, userId string, chargeUserBalance bool) {
	if preConsumedQuota == 0 {
		return
	}
	go func(ctx context.Context) {
		if strings.TrimSpace(tokenId) != "" {
			var err error
			if chargeUserBalance {
				err = model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
			} else {
				err = model.PostConsumeTokenRemainQuota(tokenId, -preConsumedQuota)
			}
			if err != nil {
				logger.Error(ctx, "error return pre-consumed quota: "+err.Error())
				emitGroupBillingFailureCard(ctx, "return_pre_consumed_quota_failed", "RETURN_PRE_CONSUMED_QUOTA_FAILED", err.Error(), userId, "", "", "", tokenId, -preConsumedQuota, preConsumedQuota, "single_user", "预扣额度回滚失败，用户或令牌可用额度可能异常")
			}
			return
		}
		if !chargeUserBalance {
			return
		}
		// JWT 场景：只需要归还用户额度
		err := model.IncreaseUserQuota(userId, preConsumedQuota)
		if err != nil {
			logger.Error(ctx, "error return pre-consumed user quota: "+err.Error())
			emitGroupBillingFailureCard(ctx, "return_pre_consumed_user_quota_failed", "RETURN_PRE_CONSUMED_USER_QUOTA_FAILED", err.Error(), userId, "", "", "", "", -preConsumedQuota, preConsumedQuota, "single_user", "用户预扣额度回滚失败，余额可能异常")
			return
		}
		_ = model.CacheUpdateUserQuota(ctx, userId)
	}(ctx)
}

func PostConsumeQuota(ctx context.Context, tokenId string, quotaDelta int64, totalQuota int64, userId string, groupID string, channelId string, pricing model.ResolvedModelPricing, groupRatio float64, modelName string, tokenName string, chargeUserBalance bool, packageReservation model.PackageQuotaReservation, snapshot BillingSnapshot) {
	// quotaDelta is remaining quota to be consumed
	var err error
	if strings.TrimSpace(tokenId) != "" {
		if chargeUserBalance {
			err = model.PostConsumeTokenQuota(tokenId, quotaDelta)
		} else {
			err = model.PostConsumeTokenRemainQuota(tokenId, quotaDelta)
		}
		if err != nil {
			logger.SysError("error consuming token remain quota: " + err.Error())
			emitGroupBillingFailureCard(ctx, "post_consume_token_quota_failed", "POST_CONSUME_TOKEN_QUOTA_FAILED", err.Error(), userId, groupID, channelId, modelName, tokenId, quotaDelta, totalQuota, "single_user", "令牌后扣费失败，账务可能不一致")
		}
	} else if chargeUserBalance {
		if quotaDelta > 0 {
			err = model.DecreaseUserQuota(userId, quotaDelta)
		} else if quotaDelta < 0 {
			err = model.IncreaseUserQuota(userId, -quotaDelta)
		}
		if err != nil {
			logger.SysError("error consuming user quota: " + err.Error())
			emitGroupBillingFailureCard(ctx, "post_consume_user_quota_failed", "POST_CONSUME_USER_QUOTA_FAILED", err.Error(), userId, groupID, channelId, modelName, tokenId, quotaDelta, totalQuota, "single_user", "用户后扣费失败，账务可能不一致")
		}
	}
	if chargeUserBalance {
		err = model.CacheUpdateUserQuota(ctx, userId)
		if err != nil {
			logger.SysError("error update user quota cache: " + err.Error())
			emitGroupBillingFailureCard(ctx, "update_user_quota_cache_failed", "UPDATE_USER_QUOTA_CACHE_FAILED", err.Error(), userId, groupID, channelId, modelName, tokenId, quotaDelta, totalQuota, "single_user", "用户额度缓存刷新失败，展示额度可能延迟")
		}
		if totalQuota > 0 {
			consumedFromLots, consumeErr := model.ConsumeUserBalanceLots(userId, totalQuota)
			if consumeErr != nil {
				logger.Error(ctx, "error consuming user balance lots: "+consumeErr.Error())
			} else if consumedFromLots < totalQuota {
				logger.Warnf(ctx, "user balance lot coverage partial user=%s consumed=%d requested=%d", strings.TrimSpace(userId), consumedFromLots, totalQuota)
			}
		}
	}
	userDailyQuota := 0
	userEmergencyQuota := 0
	if !chargeUserBalance {
		dailyConsumed, emergencyConsumed, settleErr := model.SettlePackageQuotaReservation(packageReservation, totalQuota)
		if settleErr != nil {
			logger.Error(ctx, "settle package quota reservation failed: "+settleErr.Error())
			emitGroupBillingFailureCard(ctx, "settle_package_quota_reservation_failed", "SETTLE_PACKAGE_QUOTA_RESERVATION_FAILED", settleErr.Error(), userId, groupID, channelId, modelName, tokenId, quotaDelta, totalQuota, "single_user", "分组套餐额度结算失败，账务与配额可能不一致")
		} else {
			userDailyQuota = int(dailyConsumed)
			userEmergencyQuota = int(emergencyConsumed)
		}
	}
	// totalQuota is total quota consumed
	if totalQuota != 0 {
		snapshot.YYCAmount = totalQuota
		entry := &model.Log{
			UserId:             userId,
			GroupId:            groupID,
			ChannelId:          channelId,
			PromptTokens:       int(totalQuota),
			CompletionTokens:   0,
			ModelName:          modelName,
			TokenName:          tokenName,
			Quota:              int(totalQuota),
			BillingSource:      model.ResolveConsumeLogBillingSource(chargeUserBalance),
			UserDailyQuota:     userDailyQuota,
			UserEmergencyQuota: userEmergencyQuota,
			Content:            FormatPricingLog(pricing, groupRatio),
		}
		snapshot.ApplyToLog(entry)
		model.RecordConsumeLog(ctx, entry)
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
}

func emitGroupBillingFailureCard(
	ctx context.Context,
	subtype string,
	errorCode string,
	message string,
	userID string,
	groupID string,
	channelID string,
	modelName string,
	tokenID string,
	quotaDelta int64,
	totalQuota int64,
	impactScope string,
	impactSummary string,
) {
	tags := map[string]string{}
	if strings.TrimSpace(tokenID) != "" {
		tags["token_id"] = strings.TrimSpace(tokenID)
	}
	if quotaDelta != 0 {
		tags["quota_delta"] = strconv.FormatInt(quotaDelta, 10)
	}
	if totalQuota != 0 {
		tags["total_quota"] = strconv.FormatInt(totalQuota, 10)
	}
	logger.EmitFeishuCardError(ctx, logger.ErrorCardEvent{
		EventType:     "group_billing_post_consume_error",
		Domain:        "group_billing",
		Subtype:       strings.TrimSpace(subtype),
		Severity:      "error",
		Title:         "分组计费失败",
		Summary:       strings.TrimSpace(message),
		BizStatus:     "failed",
		ErrorCode:     strings.TrimSpace(errorCode),
		ErrorMessage:  strings.TrimSpace(message),
		ImpactScope:   strings.TrimSpace(impactScope),
		ImpactSummary: strings.TrimSpace(impactSummary),
		UserID:        strings.TrimSpace(userID),
		GroupID:       strings.TrimSpace(groupID),
		ChannelID:     strings.TrimSpace(channelID),
		ModelName:     strings.TrimSpace(modelName),
		Tags:          tags,
	})
}
