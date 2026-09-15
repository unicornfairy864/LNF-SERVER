package initialization

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func Viper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/")
	
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("fail to read config.yaml: %s", err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed")
		if err := v.Unmarshal(&global.LNF_CONFIG); err != nil {
			fmt.Println(err)
		}
	})

	if err := v.Unmarshal(&global.LNF_CONFIG); err != nil {
		fmt.Println(err)
	}
	return v
}