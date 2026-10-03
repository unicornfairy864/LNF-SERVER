package handler

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/client"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/service"
	csutils "github.com/unicornfairy864/LNF-SERVER/chensong/internal/utils"
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
	utils.LogJson(req)
	if req.PostType != "message" {
		response.Success(c)
		return
	}
	utils.LogJson("ServiceStart")
	// Ping
	if req.RawMessage == "🐏" {
		_, err = client.Client.SendGroupMessage("🐏：咩。", req.GroupID)
		if err != nil {
			response.Fail(c)
			return
		}
	}
	// LNF Agent 链路（批次 3）：仅 activated_group + @机器人 + （关键词命中 或 进行中会话）
	// 命中后本条不再走 emoji 链路；触发判定轻量同步，真正的处理放 goroutine，避免阻塞 webhook
	if text := csutils.CleanForAgent(req); service.LnfAgent.Trigger(req, text) {
		if service.LnfAgent.MarkOnce(req.MessageID) {
			utils.LogJson("AgentTriggered")
			go service.LnfAgent.Handle(req, text)
		} else {
			utils.LogJson("AgentDuplicateSkipped")
		}
		response.Success(c)
		return
	}
	// TranslateEmoji
	utils.LogJson("TranslateEmojiStart")
	code = service.TranslateEmoji(req)
	if code == response.CodeSuccess {
		utils.LogJson("SuccessToTranslateEmoji")
	} else {
		utils.LogJson(strconv.Itoa(int(code)) + " " + response.Msg[code])
		response.Fail(c)
		return
	}
	//

	utils.LogJson("ServiceEnd")
	response.Success(c)
}
