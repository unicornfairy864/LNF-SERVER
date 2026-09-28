package initialization

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func InitOpenAI() {
	client := openai.NewClient(
		option.WithBaseURL(global.LNF_CONFIG.OpenAI.OpenaiBaseUrl),
		option.WithAPIKey(global.LNF_CONFIG.OpenAI.OpenaiKey),
	)
	global.LNF_OpenAI = &client
}
