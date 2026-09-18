package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		tokenString := auth[7:]
		if tokenString == "" {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		// 解析 JWT
		claims, err := utils.JWT.ParseToken(tokenString)
		if err != nil {
			switch err.Error() {
			case "InvalidSigningMethod":
				response.FailWithCode(c, response.CodeInvalidSigningMethod)
				c.Abort()
				return
			case "TokenInvalid":
				response.FailWithCode(c, response.CodeUnauthorized)
				c.Abort()
				return
			default:
				response.FailWithCode(c, response.CodeDatabaseError)
				c.Abort()
				return
			}
		}
		// JWT 版本是否正确
		var expectedVersion int64
		val, err := dao.RedisDao.GetKey("jwt:user:" + claims.Subject + ":version")
		if err != nil {
			response.FailWithCode(c, response.CodeServerError)
			c.Abort()
			return
		}
		switch v := val.(type) {
		case int64:
			expectedVersion = v
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				response.FailWithCode(c, response.CodeServerError)
				c.Abort()
				return
			}
			expectedVersion = n
		case []byte:
			n, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				response.FailWithCode(c, response.CodeServerError)
				c.Abort()
				return
			}
			expectedVersion = n
		default:
			expectedVersion = 0
		}
		if claims.JWTVersion != expectedVersion {
			response.FailWithCode(c, response.CodeTokenBanned)
			c.Abort()
			return
		}
		// JWT 是否进入黑名单
		if err := utils.JWT.IsTokenBanned(claims.ID); err != nil {
			switch err.Error() {
			case "TokenBanned":
				response.FailWithCode(c, response.CodeTokenBanned)
				c.Abort()
				return
			default:
				response.FailWithCode(c, response.CodeDatabaseError)
				c.Abort()
				return
			}
		}
		// 判断刷新 JWT
		if time.Now().After(claims.FreshAfter.Time) {
			if err := utils.JWT.BanToken(claims); err != nil {
				response.FailWithCode(c, response.CodeDatabaseError)
				c.Abort()
				return
			}
			uid, _ := strconv.ParseInt(claims.Subject, 10, 64)
			newToken, err := utils.JWT.GenerateToken(&model.User{
				ID:   uid,
				Role: claims.Role,
			}, true)
			if err != nil {
				response.FailWithCode(c, response.CodeServerError)
				c.Abort()
				return
			}
			c.Set("TokenFresh", true)
			c.Header("Authorization", "Bearer "+newToken)

			c.Set("JwtId", claims.Subject)
			c.Set("JwtRole", claims.Role)
		} else {
			c.Set("TokenFresh", false)
		}
		c.Next()
	}
}
