package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

func ReceiveMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 Header
		auth := c.Request.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		tokenString := auth[7:]
		if tokenString != global.LNF_CONFIG.ChenSong.ReceiveToken {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		c.Next()
		return
	}
}
