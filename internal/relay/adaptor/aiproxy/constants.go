package aiproxy

import "github.com/yeying-community/router/internal/relay/adaptor/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
