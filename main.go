package main

import (
	"fmt"
	"log"

	"github.com/unicornfairy864/LNF-SERVER/core"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func main() {
	var err error
	// Viper
	global.LNF_VP = core.Viper()
	// Database
	global.LNF_DB, err = core.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	tester()
}

func tester() {
	fmt.Println("This is the tester.")

	fmt.Println(global.LNF_CONFIG.Mysql)
}