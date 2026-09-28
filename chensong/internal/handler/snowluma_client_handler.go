package handler

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/service"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type SlHandler struct{}

func (sl *SlHandler) ReceiverHandler(c *gin.Context) {
	var req model.GroupMessageEvent
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 将 body 原封不动保存到 项目根/chensong/logs/chensong_log_<time>.txt
	// 运行目录即项目根（config.yaml 同样从 "." 读取），故使用相对路径
	logDir := filepath.Join("chensong", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Printf("mkdir %s failed: %v", logDir, err)
		response.Fail(c)
		return
	}
	// 文件名不含 ":" 等非法字符，Linux/Windows 均可用；追加纳秒避免同一秒内互相覆盖
	fileName := "chensong_log_" + time.Now().Format("20060102_150405_000000000") + ".txt"
	jsonReq, err := json.Marshal(req)
	if err := os.WriteFile(filepath.Join(logDir, fileName), jsonReq, 0o644); err != nil {
		response.Fail(c)
		return
	}
	// Services
	var code response.Code
	// TranslateEmoji
	code = service.TranslateEmoji(req)
	if code == response.CodeSuccess {
		utils.LogJson("Success")
	} else {
		utils.LogJson(strconv.Itoa(int(code)) + " " + response.Msg[code])
	}
	//
	response.Success(c)
}
