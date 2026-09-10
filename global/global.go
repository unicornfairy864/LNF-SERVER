package global

import (
	"github.com/spf13/viper"
	"github.com/unicornfairy864/LNF-SERVER/config"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

var (
	LNF_CONFIG config.Config
	LNF_SERVERI service.Service
	LNF_VP *viper.Viper
)