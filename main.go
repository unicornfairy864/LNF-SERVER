package main

import (
	"fmt"

	"github.com/unicornfairy864/LNF-SERVER/core"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func main() {
	global.LNF_VP = core.Viper()
	tester()
}

func tester() {
	fmt.Println("This is the tester.")

	fmt.Println(global.LNF_CONFIG.Api)
}