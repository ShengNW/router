package channel

import (
	"github.com/yeying-community/router/internal/admin/model"
	channelrepo "github.com/yeying-community/router/internal/admin/repository/channel"
)

func GetAll(start, num int, status string) ([]*model.Channel, error) {
	return channelrepo.GetAll(start, num, status)
}

func Search(keyword string) ([]*model.Channel, error) {
	return channelrepo.Search(keyword)
}

func GetByID(id int, selectAll bool) (*model.Channel, error) {
	return channelrepo.GetByID(id, selectAll)
}

func BatchInsert(channels []model.Channel) error {
	return channelrepo.BatchInsert(channels)
}

func DeleteByID(id int) error {
	return channelrepo.DeleteByID(id)
}

func DeleteDisabled() (int64, error) {
	return channelrepo.DeleteDisabled()
}

func Update(channel *model.Channel) error {
	return channelrepo.Update(channel)
}
