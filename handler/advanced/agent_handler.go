package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/agent/orchestrator"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type AgentHandlerGroup struct{}

// AgentChatHandler Agent 会话主入口
// @Summary      Agent 会话主入口（匹配 / 建帖草稿 / 确认发布）
// @Description  普通用户（role=0）的自然语言入口，会话状态存 Redis（TTL 见 openai.agent_session_ttl，滑动续期）。<br />
// @Description  首轮：自动判类 + 抽取，分为三种走向：<br />
// @Description  1) 建帖（失物/招领）→ 返回 need_confirm 草稿，回复中同时给出缺失项与确认指引（**仅一次**）；<br />
// @Description  2) 匹配 → 召回 + LLM 精排，返回 matched（含 score/reasons）或 no_match；<br />
// @Description  3) 闲聊/无关 → chitchat 固定话术。<br />
// @Description  确认轮：用户回复补充信息（自动合并后直接建帖）、回复「确认」（用草稿建帖，缺地点自动填 agent_default_location_id）、回复「取消」则结束。<br />
// @Description  建帖复用 POST /item/create 的全部校验；响应 created_item_id 为新建物品 ID。<br />
// @Description  action=confirm/cancel 可显式确认或放弃；重复确认幂等（返回同一 created_item_id）。<br />
// @Description  错误码：120003 输入非法、120004 频率超限、120005 会话状态不允许该操作、120006 功能未开启、120001 会话不存在、120002 智能服务不可用
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        request  body  model.AgentChatRequest  true  "会话请求体"
// @Success      200  {object}  response.CommonResponse{data=model.AgentChatResponse}
// @Router       /api/v1/agent/chat [post]
func (h *AgentHandlerGroup) AgentChatHandler(c *gin.Context) {
	req := model.AgentChatRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := orchestrator.Service.Chat(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// AgentMatchHandler 无状态匹配（只读）
// @Summary      Agent 匹配帖子（只读，无会话）
// @Description  按自然语言描述（可选图片）召回并精排相似帖子：主结果=相反类型（失物↔招领），相似区=同类型最多 2 条。<br />
// @Description  返回 entities 为抽取到的结构（前端可用于回显），matches[].item 为完整 ItemResponse（与详情接口口径一致）。<br />
// @Description  verdict：strong_match 强匹配 / ambiguous 模糊 / no_match 无匹配（无匹配时仍返回 code=0，matches 为空数组）。
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        request  body  model.AgentMatchRequest  true  "匹配请求体"
// @Success      200  {object}  response.CommonResponse{data=model.AgentMatchResponse}
// @Router       /api/v1/agent/match [post]
func (h *AgentHandlerGroup) AgentMatchHandler(c *gin.Context) {
	req := model.AgentMatchRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := orchestrator.Service.Match(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// AgentExtractHandler 无状态抽取（只读）
// @Summary      Agent 信息抽取（只读，智能填充表单）
// @Description  只做「语境判定 + 意图识别 + 信息抽取」，不检索、不建帖。<br />
// @Description  draft.tag_ids / location_id / lost_found_time 可直接提交给 POST /item/create；<br />
// @Description  不在词表内的标签与地点会被丢弃（置空），模型编造的字段一律不采信。
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        request  body  model.AgentExtractRequest  true  "抽取请求体"
// @Success      200  {object}  response.CommonResponse{data=model.AgentExtractResponse}
// @Router       /api/v1/agent/extract [post]
func (h *AgentHandlerGroup) AgentExtractHandler(c *gin.Context) {
	req := model.AgentExtractRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := orchestrator.Service.Extract(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// AgentSessionCloseHandler 关闭会话
// @Summary      关闭 Agent 会话
// @Description  清理当前用户的会话（幂等）；会话不存在时同样返回成功
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        request  body  model.AgentSessionCloseRequest  false  "关闭会话请求体（可省略）"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/agent/session/close [post]
func (h *AgentHandlerGroup) AgentSessionCloseHandler(c *gin.Context) {
	req := model.AgentSessionCloseRequest{}
	// 请求体可省略：绑定失败不视为错误
	_ = c.ShouldBindBodyWithJSON(&req)
	code := orchestrator.Service.CloseSession(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}
