package token

import (
	"github.com/yeying-community/router/internal/admin/model"
	tokenrepo "github.com/yeying-community/router/internal/admin/repository/token"
)

func GetAll(userId, start, num int, order string) ([]*model.Token, error) {
	return tokenrepo.GetAll(userId, start, num, order)
}

func Search(userId int, keyword string) ([]*model.Token, error) {
	return tokenrepo.Search(userId, keyword)
}

func GetByIDs(tokenId, userId int) (*model.Token, error) {
	return tokenrepo.GetByIDs(tokenId, userId)
}

func GetByID(tokenId int) (*model.Token, error) {
	return tokenrepo.GetByID(tokenId)
}

func Create(token *model.Token) error {
	return tokenrepo.Create(token)
}

func Update(token *model.Token) error {
	return tokenrepo.Update(token)
}

func DeleteByID(tokenId, userId int) error {
	return tokenrepo.DeleteByID(tokenId, userId)
}
