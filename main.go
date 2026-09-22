package main

import (
	"fmt"
	"log"

	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/initialization"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 空导入以执行 docs 包的 init() 函数，向 ginSwagger 注册 Swagger 路由
	_ "github.com/unicornfairy864/LNF-SERVER/docs"
)

func main() {
	// Viper
	global.LNF_VP = initialization.Viper()

	// Database
	var err error
	global.LNF_DB, err = initialization.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer initialization.CloseDB()

	// Redis
	global.LNF_RDB, err = initialization.InitRedis()
	if err != nil {
		log.Fatal(err)
	}
	defer global.LNF_RDB.Close()

	// Http Client
	global.LNF_Resty = initialization.InitResty()

	// Router
	r := initialization.InitRouter()

	// Swagger
	r.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	test()

	// Start server
	r.Run(fmt.Sprintf(":%d", global.LNF_CONFIG.Server.Port))
}

func test() {
	fmt.Print("这是测试")
}
