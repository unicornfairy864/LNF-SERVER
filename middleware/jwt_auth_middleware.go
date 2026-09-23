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

var (
	ContextID       = "jwt:id"
	ContextNickname = "jwt:nickname"
	ContextRole     = "jwt:role"

	ContextJti       = "jwt:jti"
	ContextExpiredAt = "jwt:expired_at"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 Header
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
		expectedVersion, err := dao.RedisDao.GetValueInt64("jwt:user:" + claims.Subject + ":version")
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
			}, false)

			if err != nil {
				response.FailWithCode(c, response.CodeServerError)
				c.Abort()
				return
			}
			c.Set("TokenFresh", true)
			c.Header("Authorization", "Bearer "+newToken)
		} else {
			c.Set("TokenFresh", false)
		}

		uid, _ := strconv.ParseInt(claims.Subject, 10, 64)
		c.Set(ContextID, uid)
		c.Set(ContextNickname, claims.Nickname)
		c.Set(ContextRole, claims.Role)

		c.Set(ContextJti, claims.ID)
		c.Set(ContextExpiredAt, claims.ExpiresAt.Time)
		c.Next()
		return
	}
}

func SystemAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetInt8("jwt:role")
		utils.LogJson(role)
		if role != 2 {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		c.Next()
		return
	}
}

func ServiceAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetInt8("jwt:role")
		if role == 0 {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		c.Next()
		return
	}
}
