package handler

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type SlHandler struct{}

func (sl *SlHandler) ReceiverHandler(c *gin.Context) {
	var body string
	if err := c.ShouldBindBodyWithPlain(&body); err != nil {
		response.Fail(c)
		return
	}
	// 保存到 logs/received/当前时间.json
	dir := filepath.Join("logs", "received")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.Fail(c)
		return
	}
	path := filepath.Join(dir, time.Now().Format("2006-01-02_15-04-05.000")+".json")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		response.Fail(c)
		return
	}
	// Receive Service
	response.Success(c)
}
