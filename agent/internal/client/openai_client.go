package client

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

type Client struct{}

func (c *Client) EasyRequest(system, user string) (string, error) {
	ctx := context.Background()

	params := openai.ChatCompletionNewParams{
		Model: global.LNF_CONFIG.OpenAI.DefaultModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		},
		Temperature: openai.Float(global.LNF_CONFIG.OpenAI.Temperature),
	}

	resp, err := global.LNF_OpenAI.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Choices[0].Message.Content, nil
}
