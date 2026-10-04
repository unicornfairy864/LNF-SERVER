package utils

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
)

type MessageSegment struct {
	Type string `json:"type"`
	Data struct {
		Text string `json:"text"`
	} `json:"data"`
}

// CQ 码内的 [ ] 会被转义成 &#91;&#93;，所以码内不会出现 ]，[^\]]* 安全
var cqRegex = regexp.MustCompile(`\[CQ:[^\]]*\]`)

// CleanEvent 输入完整事件 JSON，输出纯文本；空串 = 无文本，直接跳过不调模型
func CleanEvent(e model.GroupMessageEvent) string {
	return extractText(e.Message, e.RawMessage)
}

func extractText(msg json.RawMessage, rawMessage string) string {
	s := strings.TrimSpace(string(msg))
	switch {
	case s == "" || s == "null": // message 缺失 → 用 raw_message
		return stripCQ(rawMessage)
	case s[0] == '[': // 数组格式：只留 text 段，其余段全丢
		var segs []MessageSegment
		if err := json.Unmarshal(msg, &segs); err == nil {
			var b strings.Builder
			for _, seg := range segs {
				if seg.Type == "text" {
					b.WriteString(seg.Data.Text)
				}
			}
			return strings.TrimSpace(stripCQ(b.String()))
		}
	case s[0] == '"': // CQ 字符串格式：删掉所有 CQ 码
		var str string
		if err := json.Unmarshal(msg, &str); err == nil {
			return stripCQ(str)
		}
	}
	return stripCQ(rawMessage) // 解析失败兜底
}

func stripCQ(s string) string {
	return strings.TrimSpace(cqRegex.ReplaceAllString(s, ""))
}

// ==================== Agent 专用清洗（emoji 链路的 CleanEvent 保持不动） ====================

// cqMediaRegex 媒体类 CQ 码（图片/语音/视频/文件/表情等；图片段常带超长 URI）整体剥离
var cqMediaRegex = regexp.MustCompile(`\[CQ:(image|record|video|file|face|mface|rps|dice|json|xml|markdown)[^\]]*\]`)

// urlRegex http(s) 链接（含图片段残留的长 URI）
var urlRegex = regexp.MustCompile(`https?://[^\s\]]+`)

// cqAtRegex CQ 字符串格式的 @ 段（用于判断是否 @ 了机器人）
var cqAtRegex = regexp.MustCompile(`\[CQ:at,qq=([0-9]+)[^\]]*\]`)

// agentSegment 数组格式的消息段（Data 原样保留，由各段自行解析）
type agentSegment struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// CleanForAgent 给 Agent 链路用的清洗：剥离媒体段/CQ 码/URL，返回纯文本
// 与 CleanEvent 的区别：显式处理数组与 CQ 字符串两种格式，并清掉图片段里的长 URI
func CleanForAgent(e model.GroupMessageEvent) string {
	raw := extractTextForAgent(e)
	raw = cqMediaRegex.ReplaceAllString(raw, " ")
	raw = urlRegex.ReplaceAllString(raw, " ")
	raw = cqRegex.ReplaceAllString(raw, " ") // 其余 CQ 码（[CQ:at]、[CQ:reply] 等）
	return strings.Join(strings.Fields(raw), " ")
}

// extractTextForAgent 取消息文本：数组格式只留 text 段；字符串/CQ 格式原样返回
func extractTextForAgent(e model.GroupMessageEvent) string {
	s := strings.TrimSpace(string(e.Message))
	switch {
	case s == "" || s == "null":
		return e.RawMessage
	case s[0] == '[':
		var segs []agentSegment
		if err := json.Unmarshal(e.Message, &segs); err == nil {
			var b strings.Builder
			for _, seg := range segs {
				if seg.Type != "text" {
					continue
				}
				var d struct {
					Text string `json:"text"`
				}
				if err := json.Unmarshal(seg.Data, &d); err == nil {
					b.WriteString(d.Text)
					b.WriteString(" ")
				}
			}
			if b.Len() > 0 {
				return b.String()
			}
		}
		return e.RawMessage
	case s[0] == '"':
		var str string
		if err := json.Unmarshal(e.Message, &str); err == nil {
			return str
		}
	}
	return e.RawMessage
}

// HasAtBot 判断消息是否 @ 了指定 QQ（兼容数组 at 段与 CQ 字符串；不查库）
func HasAtBot(e model.GroupMessageEvent, botQQ int64) bool {
	if botQQ <= 0 {
		return false
	}
	target := strconv.FormatInt(botQQ, 10)
	s := strings.TrimSpace(string(e.Message))
	if strings.HasPrefix(s, "[") {
		var segs []agentSegment
		if err := json.Unmarshal(e.Message, &segs); err == nil {
			for _, seg := range segs {
				if seg.Type != "at" {
					continue
				}
				if qq, ok := segmentAtQQ(seg.Data); ok && qq == target {
					return true
				}
			}
		}
	}
	for _, m := range cqAtRegex.FindAllStringSubmatch(e.RawMessage, -1) {
		if len(m) > 1 && m[1] == target {
			return true
		}
	}
	if strings.HasPrefix(s, "\"") {
		var str string
		if err := json.Unmarshal(e.Message, &str); err == nil {
			for _, m := range cqAtRegex.FindAllStringSubmatch(str, -1) {
				if len(m) > 1 && m[1] == target {
					return true
				}
			}
		}
	}
	return false
}

// segmentAtQQ 解析 at 段的 qq 字段（兼容字符串与数字两种写法）
func segmentAtQQ(raw json.RawMessage) (string, bool) {
	var m struct {
		QQ json.RawMessage `json:"qq"`
	}
	if err := json.Unmarshal(raw, &m); err != nil || len(m.QQ) == 0 {
		return "", false
	}
	v := strings.Trim(strings.TrimSpace(string(m.QQ)), "\"")
	if v == "" {
		return "", false
	}
	return v, true
}

// ContainsKeyword 关键词匹配（已废弃：2026-10-04 起触发不再依赖关键词，保留工具函数备用）
func ContainsKeyword(text, keywords string) bool {
	text = strings.ToLower(text)
	for _, kw := range strings.Split(keywords, ",") {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw == "" {
			continue
		}
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
