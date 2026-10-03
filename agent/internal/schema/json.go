package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidJSON 模型输出无法解析为 JSON 对象
var ErrInvalidJSON = errors.New("schema: invalid json")

// ParseJSONObject 从模型原始输出中提取首个完整 JSON 对象。
// 兼容：Markdown 围栏（```json）、前后解释性文字、BOM/空白；括号配对扫描考虑字符串与转义。
func ParseJSONObject(raw string) (string, error) {
	s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "\ufeff"))
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i >= 0 {
			s = s[i+1:]
		}
		if j := strings.LastIndex(s, "```"); j > 0 {
			s = s[:j]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.Index(s, "{")
	if start < 0 {
		return "", fmt.Errorf("%w: no object found", ErrInvalidJSON)
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case ch == '\\':
				esc = true
			case ch == '"':
				inStr = false
			}
			continue
		}
		switch ch {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				candidate := s[start : i+1]
				if !json.Valid([]byte(candidate)) {
					return "", fmt.Errorf("%w: object not valid", ErrInvalidJSON)
				}
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("%w: unterminated object", ErrInvalidJSON)
}

// UnmarshalObject 解析模型输出并解码到目标结构（未知字段自动忽略 = 字段白名单）
func UnmarshalObject(raw string, dst any) error {
	obj, err := ParseJSONObject(raw)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(obj), dst); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return nil
}

// ==================== 通用清洗工具 ====================

// truncateRunes 按字符数截断（不会切坏 UTF-8）
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// truncateBytes 按字节上限截断（对齐 SQL 列宽，如 varchar(100)），保证不切坏 UTF-8
func truncateBytes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}
	return s[:cut]
}

// utf8Start 判断字节是否为 UTF-8 起始字节
func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

// cleanStr 去首尾空白
func cleanStr(s string) string { return strings.TrimSpace(s) }

// cleanStrPtr 去空白并转指针；空串返回 nil
func cleanStrPtr(s string) *string {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return &t
}

// cleanStrings 去重、去空、限长、限量
func cleanStrings(in []string, maxItems, maxRunes int) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, v := range in {
		v = truncateRunes(strings.TrimSpace(v), maxRunes)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		if len(out) >= maxItems {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
