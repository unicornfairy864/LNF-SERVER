package client

import (
	"encoding/json"
	"fmt"

	"github.com/unicornfairy864/LNF-SERVER/global"
)

type ChenSongClient struct{}

type NapcatRequest struct {
	// Basic
	UserID    int64  `json:"user_id,omitempty"`
	GroupID   int64  `json:"group_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	// Extra
	Message []OB11Message `json:"message,omitempty"`
}

type NapcatResponse struct {
	Status  string          `json:"status"`
	Retcode int             `json:"retcode"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Wording string          `json:"wording"`
	Echo    string          `json:"echo"`
	Stream  string          `json:"stream"`
}

func (c *ChenSongClient) Request(method string, url string, napcatReq *NapcatRequest) (*NapcatResponse, error) {
	req, err := global.LNF_Resty.R().
		SetHeader("Authorization", "Bearer "+global.LNF_CONFIG.ChenSong.ApiToken).
		SetBody(napcatReq).
		Execute(method, url)
	if err != nil {
		return nil, err
	}
	if req.IsError() {
		return nil, fmt.Errorf("HttpRequestError")
	}
	var result NapcatResponse
	if err := json.Unmarshal(req.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析 napcat 响应失败: %w", err)
	}
	return &result, nil
}
