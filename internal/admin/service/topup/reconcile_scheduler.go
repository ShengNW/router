package topup

import (
	"strings"
	"sync"
	"time"

	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
)

const (
	topupReconcileLoopIntervalSeconds = 45
	topupReconcileCooldownSeconds     = 30
	topupReconcileBatchSize           = 20
)

var startTopupReconcileWorkerOnce sync.Once

func StartTopupReconcileWorker() {
	startTopupReconcileWorkerOnce.Do(func() {
		go runTopupReconcileWorker()
	})
}

func runTopupReconcileWorker() {
	logger.SysLog("[topup.reconcile] worker started")
	ticker := time.NewTicker(topupReconcileLoopIntervalSeconds * time.Second)
	defer ticker.Stop()

	for {
		runTopupReconcileOnce()
		<-ticker.C
	}
}

func runTopupReconcileOnce() {
	maxUpdatedAt := helper.GetTimestamp() - topupReconcileCooldownSeconds
	rows, err := model.ListTopupOrderReconcileCandidatesWithDB(model.DB, topupReconcileBatchSize, maxUpdatedAt)
	if err != nil {
		logger.SysWarnf("[topup.reconcile] list candidates failed: %s", err.Error())
		return
	}
	if len(rows) == 0 {
		return
	}

	successCount := 0
	failedCount := 0
	for _, row := range rows {
		if strings.TrimSpace(row.Id) == "" || strings.TrimSpace(row.UserID) == "" {
			continue
		}
		if _, err := model.RefreshTopupOrderStatusWithDB(model.DB, row.Id, row.UserID); err != nil {
			failedCount++
			logger.SysWarnf("[topup.reconcile] refresh failed order_id=%s status=%s err=%s",
				strings.TrimSpace(row.Id),
				strings.TrimSpace(row.Status),
				err.Error(),
			)
			continue
		}
		successCount++
	}
	if successCount > 0 || failedCount > 0 {
		logger.SysLogf("[topup.reconcile] batch finished success=%d failed=%d", successCount, failedCount)
	}
}
