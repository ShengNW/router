package log

import "github.com/yeying-community/router/internal/admin/model"

func GetAll(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int) ([]*model.Log, error) {
	return model.GetAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, startIdx, num, channel)
}

func GetUser(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int) ([]*model.Log, error) {
	return model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, startIdx, num)
}

func SearchAll(keyword string) ([]*model.Log, error) {
	return model.SearchAllLogs(keyword)
}

func SearchUser(userId int, keyword string) ([]*model.Log, error) {
	return model.SearchUserLogs(userId, keyword)
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) int64 {
	return model.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel)
}

func DeleteOld(targetTimestamp int64) (int64, error) {
	return model.DeleteOldLog(targetTimestamp)
}
