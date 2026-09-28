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
	// 将 body 原封不动保存到 项目根/chensong/logs/chensong_log_<time>.txt
	// 运行目录即项目根（config.yaml 同样从 "." 读取），故使用相对路径
	logDir := filepath.Join("chensong", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		response.Fail(c)
		return
	}
	// 文件名不含 ":" 等非法字符，Linux/Windows 均可用；追加纳秒避免同一秒内互相覆盖
	fileName := "chensong_log_" + time.Now().Format("20060102_150405_000000000") + ".txt"
	if err := os.WriteFile(filepath.Join(logDir, fileName), []byte(body), 0o644); err != nil {
		response.Fail(c)
		return
	}
	response.Success(c)
}
