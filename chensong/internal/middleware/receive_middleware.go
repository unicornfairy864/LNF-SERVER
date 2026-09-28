package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type SlMiddleware struct{}

func (s *SlMiddleware) ReceiveMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		sig := c.Request.Header.Get("X-Signature")
		if !strings.HasPrefix(sig, "sha1=") {
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		gotHex := strings.TrimPrefix(sig, "sha1=")

		// 读原始 body —— 注意 HMAC 必须基于原始字节，不能是反序列化后的
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("[chensong] read body failed: %v", err)
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		// 关键：读完后写回，否则后面的 handler 读不到 body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 用配置的 secret 算 HMAC-SHA1
		mac := hmac.New(sha1.New, []byte(global.LNF_CONFIG.ChenSong.ReceiveToken))
		mac.Write(bodyBytes)
		wantHex := hex.EncodeToString(mac.Sum(nil))

		// 常量时间比较，防止时序攻击
		if !hmac.Equal([]byte(gotHex), []byte(wantHex)) {
			log.Printf("[chensong] signature mismatch: got=%s want=%s", gotHex, wantHex)
			response.FailWithCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}

		log.Printf("[chensong] signature ok, %d bytes", len(bodyBytes))
		c.Next()
	}
}
