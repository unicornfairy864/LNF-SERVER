package client

import "github.com/unicornfairy864/LNF-SERVER/global"

type ChenSongClient struct{}

type NapcatRequest struct {
	// Basic
	UserID    int64  `json:"user_id,omitempty"`
	GroupID   int64  `json:"group_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	// Extra
	Message []OB11Message `json:"message,omitempty"`
}

func (c *ChenSongClient) Request(method string, url string, napcatReq *NapcatRequest) ([]byte, error) {
	req, err := global.LNF_Resty.R().
		setURL(global.LNF_CONFIG.ChenSong.ApiUrl).
		SetMethod(method).
		SetHeader("Authorization", "Bearer "+global.LNF_CONFIG.ChenSong.ApiToken).
		SetBody(napcatReq).
		Send()
	return res, err
}
