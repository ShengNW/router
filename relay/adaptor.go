package relay

import (
	"github.com/yeying-community/router/relay/adaptor"
	"github.com/yeying-community/router/relay/adaptor/aiproxy"
	"github.com/yeying-community/router/relay/adaptor/ali"
	"github.com/yeying-community/router/relay/adaptor/anthropic"
	"github.com/yeying-community/router/relay/adaptor/aws"
	"github.com/yeying-community/router/relay/adaptor/baidu"
	"github.com/yeying-community/router/relay/adaptor/cloudflare"
	"github.com/yeying-community/router/relay/adaptor/cohere"
	"github.com/yeying-community/router/relay/adaptor/coze"
	"github.com/yeying-community/router/relay/adaptor/deepl"
	"github.com/yeying-community/router/relay/adaptor/gemini"
	"github.com/yeying-community/router/relay/adaptor/ollama"
	"github.com/yeying-community/router/relay/adaptor/openai"
	"github.com/yeying-community/router/relay/adaptor/palm"
	"github.com/yeying-community/router/relay/adaptor/proxy"
	"github.com/yeying-community/router/relay/adaptor/replicate"
	"github.com/yeying-community/router/relay/adaptor/tencent"
	"github.com/yeying-community/router/relay/adaptor/vertexai"
	"github.com/yeying-community/router/relay/adaptor/xunfei"
	"github.com/yeying-community/router/relay/adaptor/zhipu"
	"github.com/yeying-community/router/relay/apitype"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.AIProxyLibrary:
		return &aiproxy.Adaptor{}
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Anthropic:
		return &anthropic.Adaptor{}
	case apitype.AwsClaude:
		return &aws.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.Gemini:
		return &gemini.Adaptor{}
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.PaLM:
		return &palm.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Xunfei:
		return &xunfei.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.Ollama:
		return &ollama.Adaptor{}
	case apitype.Coze:
		return &coze.Adaptor{}
	case apitype.Cohere:
		return &cohere.Adaptor{}
	case apitype.Cloudflare:
		return &cloudflare.Adaptor{}
	case apitype.DeepL:
		return &deepl.Adaptor{}
	case apitype.VertexAI:
		return &vertexai.Adaptor{}
	case apitype.Proxy:
		return &proxy.Adaptor{}
	case apitype.Replicate:
		return &replicate.Adaptor{}
	}
	return nil
}
