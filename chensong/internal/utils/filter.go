package utils

import (
	"encoding/json"
	"regexp"
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
