package initialization

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/chensong"
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/router"
)

func InitRouter() (r *gin.Engine) {
	// 初始化路由
	r = gin.Default()
	api := r.Group(global.LNF_CONFIG.Server.RouterPrefix)
	{
		// Basic
		router.UserRouter.CreateRouter(api)
		// Advanced
		router.LocationRouter.CreateRouter(api)
		// ChenSong
		chensongGroup := api.Group("chensong")
		chensongGroup.Use(chensong.SlMiddleware.ReceiveMiddleWare()).POST("/receive", chensong.SlHandler.ReceiverHandler)
	}

	// 提示信息
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LNF-SERVER</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #f7f8fa;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",
                   "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
      color: #1a1a1a;
      -webkit-font-smoothing: antialiased;
    }

    .card {
      text-align: center;
      padding: 56px 48px;
      animation: fadeUp 0.6s ease both;
    }

    .title {
      font-size: 40px;
      font-weight: 600;
      letter-spacing: 4px;
      color: #111;
    }

    .divider {
      width: 40px;
      height: 2px;
      background: #d0d0d0;
      margin: 22px auto;
      border-radius: 2px;
    }

    .link {
      display: inline-block;
      font-size: 13px;
      color: #8a8f98;
      text-decoration: none;
      letter-spacing: 0.5px;
      transition: color 0.25s ease;
    }

    .link:hover {
      color: #111;
    }

    .link .path {
      border-bottom: 1px dashed #c8ccd2;
      padding-bottom: 2px;
    }

    @keyframes fadeUp {
      from {
        opacity: 0;
        transform: translateY(12px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    @media (max-width: 480px) {
      .title {
        font-size: 28px;
        letter-spacing: 2px;
      }
      .card {
        padding: 40px 24px;
      }
    }
  </style>
</head>
<body>
  <main class="card">
    <h1 class="title">LNF-SERVER</h1>
    <div class="divider"></div>
    <a class="link" href="https://github.com/Estrelas-star/lost-and-found">
      <span class="path">https://github.com/Estrelas-star/lost-and-found</span> - 访问我们的前端仓库
    </a>
	<br />
    <a class="link" href="https://github.com/unicornfairy864/LNF-SERVER">
      <span class="path">https://github.com/unicornfairy864/LNF-SERVER</span> - 访问我们的后端仓库
    </a>
    <div class="divider"></div>
    <a class="link" href="/api/v1/swagger/index.html">
      <span class="path">/api/v1/swagger/index.html</span> - 访问 swagger 文档以查看 api 接口
    </a>
	<br />
    <a class="link" href="https://github.com/unicornfairy864/LNF-SERVER/blob/main/response/response_code.go">
      <span class="path">response_code.go</span> - 在 github 仓库上查看统一错误返回码
    </a>
  </main>
</body>
</html>
		`))
	})
	return
}
