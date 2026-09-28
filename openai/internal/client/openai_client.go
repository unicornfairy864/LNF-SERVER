package client

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

type Client struct{}

func (c *Client) SendRequest(system, user string, tools ]) string {
	ctx := context.Background()
	resp, err := global.LNF_OpenAI.Responses.New(ctx, responses.ResponseNewParams{
		Model:       global.LNF_CONFIG.OpenAI.DefaultModel,
		Input:       responses.ResponseNewParamsInputUnion{},
		Temperature: openai.Float(global.LNF_CONFIG.OpenAI.Temperature),
	})
	return ""
}
