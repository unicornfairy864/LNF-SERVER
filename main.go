package main

import (
	"fmt"
	// "log"

	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/initialization"
)

func main() {
	// Viper
	global.LNF_VP = initialization.Viper()

	// Tester
	tester()

	// Database
	// var err error
	// global.LNF_DB, err = initialization.InitDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer initialization.CloseDB();

	// Router & Handler
	r := initialization.InitRouter();

	// Start server
	r.Run(fmt.Sprintf(":%d", global.LNF_CONFIG.Server.Port))
}

func tester() {
	fmt.Println("This is the tester.")

	fmt.Println(global.LNF_CONFIG.Server)
}