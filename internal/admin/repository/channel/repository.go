package channel

import "github.com/yeying-community/router/internal/admin/model"

func GetAll(start, num int, status string) ([]*model.Channel, error) {
	return model.GetAllChannels(start, num, status)
}

func Search(keyword string) ([]*model.Channel, error) {
	return model.SearchChannels(keyword)
}

func GetByID(id int, selectAll bool) (*model.Channel, error) {
	return model.GetChannelById(id, selectAll)
}

func BatchInsert(channels []model.Channel) error {
	return model.BatchInsertChannels(channels)
}

func DeleteByID(id int) error {
	channel := model.Channel{Id: id}
	return channel.Delete()
}

func DeleteDisabled() (int64, error) {
	return model.DeleteDisabledChannel()
}

func Update(channel *model.Channel) error {
	return channel.Update()
}
