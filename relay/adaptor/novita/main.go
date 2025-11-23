package novita

import (
	"fmt"

	"github.com/yeying-community/router/relay/meta"
	"github.com/yeying-community/router/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	if meta.Mode == relaymode.ChatCompletions {
		return fmt.Sprintf("%s/chat/completions", meta.BaseURL), nil
	}
	return "", fmt.Errorf("unsupported relay mode %d for novita", meta.Mode)
}
