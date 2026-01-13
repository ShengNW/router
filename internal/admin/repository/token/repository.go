package token

import "github.com/yeying-community/router/internal/admin/model"

func GetAll(userId, start, num int, order string) ([]*model.Token, error) {
	return model.GetAllUserTokens(userId, start, num, order)
}

func Search(userId int, keyword string) ([]*model.Token, error) {
	return model.SearchUserTokens(userId, keyword)
}

func GetByIDs(tokenId, userId int) (*model.Token, error) {
	return model.GetTokenByIds(tokenId, userId)
}

func GetByID(tokenId int) (*model.Token, error) {
	return model.GetTokenById(tokenId)
}

func Create(token *model.Token) error {
	return token.Insert()
}

func Update(token *model.Token) error {
	return token.Update()
}

func DeleteByID(tokenId, userId int) error {
	return model.DeleteTokenById(tokenId, userId)
}
