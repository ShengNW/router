package log

import (
	"github.com/yeying-community/router/internal/admin/model"
	logrepo "github.com/yeying-community/router/internal/admin/repository/log"
)

func GetAll(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int) ([]*model.Log, error) {
	return logrepo.GetAll(logType, startTimestamp, endTimestamp, modelName, username, tokenName, startIdx, num, channel)
}

func GetUser(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int) ([]*model.Log, error) {
	return logrepo.GetUser(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, startIdx, num)
}

func SearchAll(keyword string) ([]*model.Log, error) {
	return logrepo.SearchAll(keyword)
}

func SearchUser(userId int, keyword string) ([]*model.Log, error) {
	return logrepo.SearchUser(userId, keyword)
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) int64 {
	return logrepo.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel)
}

func DeleteOld(targetTimestamp int64) (int64, error) {
	return logrepo.DeleteOld(targetTimestamp)
}
