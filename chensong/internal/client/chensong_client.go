package client

import (
	"encoding/json"
	"fmt"

	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

type ChenSongClient struct{}

var Client = &ChenSongClient{}

func (c *ChenSongClient) Request(actionUri, method string, RequestBody interface{}) (*model.SnowLumaResponse, error) {
	req, err := global.LNF_Resty.R().
		SetHeader("Authorization", "Bearer "+global.LNF_CONFIG.ChenSong.ApiToken).
		SetBody(RequestBody).
		Execute(method, global.LNF_CONFIG.ChenSong.ApiUrl+actionUri)
	if err != nil {
		return nil, err
	}
	if req.IsError() {
		return nil, fmt.Errorf("HttpRequestError")
	}
	var result model.SnowLumaResponse
	if err := json.Unmarshal(req.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析 napcat 响应失败: %w", err)
	}
	return &result, nil
}

func (c *ChenSongClient) SendGroupMessage(msg string, group int64) (*model.SnowLumaResponse, error) {
	res, err := c.Request("/send_group_msg", "POST", model.SnowLumaSendGroupMessage{
		GroupID: group,
		Message: msg,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *ChenSongClient) SendPrivateMessage(msg string, user int64) (*model.SnowLumaResponse, error) {
	res, err := c.Request("/send_private_msg", "POST", model.SnowLumaSendPrivateMessage{
		UserID:  user,
		Message: msg,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *ChenSongClient) GetGroupMemberList() (*[]model.OB11Member, error) {
	res, err := c.Request("/get_group_member_list", "POST", struct {
		GroupID int64 `json:"group_id"`
	}{
		GroupID: global.LNF_CONFIG.ChenSong.ActivatedGroup,
	})
	if err != nil {
		return nil, err
	}
	var payload []model.OB11Member
	err = json.Unmarshal(res.Data, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}
