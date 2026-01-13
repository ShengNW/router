package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/yeying-community/router/internal/admin/model"
)

func GetAll(start, num int, order string) ([]*model.User, error) {
	return model.GetAllUsers(start, num, order)
}

func Search(keyword string) ([]*model.User, error) {
	return model.SearchUsers(keyword)
}

func GetByID(id int, selectAll bool) (*model.User, error) {
	return model.GetUserById(id, selectAll)
}

func GetByUsername(username string) (*model.User, error) {
	user := model.User{Username: username}
	err := model.DB.Where(&user).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetIDByAffCode(code string) (int, error) {
	return model.GetUserIdByAffCode(code)
}

func Create(ctx context.Context, user *model.User, inviterId int) error {
	return user.Insert(ctx, inviterId)
}

func Update(user *model.User, updatePassword bool) error {
	return user.Update(updatePassword)
}

func DeleteByID(id int) error {
	return model.DeleteUserById(id)
}

func Delete(user *model.User) error {
	return user.Delete()
}

func FillByID(user *model.User) error {
	return user.FillUserById()
}

func SearchLogsByDayAndModel(userId, start, end int) ([]*model.LogStatistic, error) {
	return model.SearchLogsByDayAndModel(userId, start, end)
}

func AccessTokenExists(token string) (bool, error) {
	var user model.User
	err := model.DB.Where("access_token = ?", token).First(&user).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func Redeem(ctx context.Context, key string, userId int) (int64, error) {
	return model.Redeem(ctx, key, userId)
}

func IncreaseQuota(userId int, quota int64) error {
	return model.IncreaseUserQuota(userId, quota)
}

func RecordLog(ctx context.Context, userId int, logType int, content string) {
	model.RecordLog(ctx, userId, logType, content)
}

func RecordTopupLog(ctx context.Context, userId int, remark string, quota int) {
	model.RecordTopupLog(ctx, userId, remark, quota)
}

func GetQuota(userId int) (int64, error) {
	return model.GetUserQuota(userId)
}

func GetUsedQuota(userId int) (int64, error) {
	return model.GetUserUsedQuota(userId)
}
