package option

import "github.com/yeying-community/router/internal/admin/model"

func Update(key string, value string) error {
	return model.UpdateOption(key, value)
}
