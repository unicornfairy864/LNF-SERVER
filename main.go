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

	// Database
	// var err error
	// global.LNF_DB, err = initialization.InitDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer initialization.CloseDB();

	// Router & Handler
	initialization.InitRouter();

	// Tester
	tester()
}

func tester() {
	fmt.Println("This is the tester.")

	// fmt.Println(global.LNF_CONFIG.Mysql)
}