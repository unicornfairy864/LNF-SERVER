package initialization

import (
	"github.com/go-resty/resty/v2"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func InitResty() *resty.Client {
	client := resty.New().
		SetTimeout(global.LNF_CONFIG.Server.ConnectionTimeout)
	return client
}
