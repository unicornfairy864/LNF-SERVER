package schema

import (
	"sort"
	"strings"
)

// ==================== 精排（rerank）结果 ====================

const (
	VerdictStrong    = "strong_match" // 强匹配（存在 score ≥ strong 阈值）
	VerdictAmbiguous = "ambiguous"    // 模糊（最高分处于 [ambiguous, strong)）
	VerdictNoMatch   = "no_match"     // 无匹配（全部 < ambiguous 阈值）
)

// IsValidVerdict 判断 verdict 是否合法
func IsValidVerdict(v string) bool {
	switch v {
	case VerdictStrong, VerdictAmbiguous, VerdictNoMatch:
		return true
	}
	return false
}

// RerankItem 单条候选的评分
type RerankItem struct {
	ItemID  int64    `json:"item_id"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
	Risk    []string `json:"risk"`
}

// NeedMoreInfo 模糊匹配时用于区分候选的追问
type NeedMoreInfo struct {
	Question string `json:"question"`
	Purpose  string `json:"purpose"`
}

// RerankResult 精排输出
type RerankResult struct {
	Ranked       []RerankItem  `json:"ranked"`
	Verdict      string        `json:"verdict"`
	NeedMoreInfo *NeedMoreInfo `json:"need_more_info"`
	Summary      string        `json:"summary"`
}

// Top 返回得分最高的候选（无则 nil）
func (r *RerankResult) Top() *RerankItem {
	if len(r.Ranked) == 0 {
		return nil
	}
	return &r.Ranked[0]
}

// TopN 返回前 n 条候选
func (r *RerankResult) TopN(n int) []RerankItem {
	if n <= 0 || len(r.Ranked) == 0 {
		return nil
	}
	if n > len(r.Ranked) {
		n = len(r.Ranked)
	}
	return r.Ranked[:n]
}

// ValidateRerank 解析并校验精排输出：
//   - 丢弃不在候选集内的 item_id（防模型编造）
//   - score 收敛到 [0,1]；按 score 降序排序；最多保留 10 条
//   - reasons ≤3 条（各 ≤20 字）；verdict 非法时按阈值推导
//   - need_more_info 仅在 verdict=ambiguous 时保留
func ValidateRerank(raw string, candidateIDs map[int64]bool, strong, ambiguous float64) (*RerankResult, error) {
	var r RerankResult
	if err := UnmarshalObject(raw, &r); err != nil {
		return nil, err
	}
	ranked := make([]RerankItem, 0, len(r.Ranked))
	seen := make(map[int64]struct{}, len(r.Ranked))
	for _, it := range r.Ranked {
		if !candidateIDs[it.ItemID] {
			continue // 模型编造的 id，丢弃
		}
		if _, dup := seen[it.ItemID]; dup {
			continue
		}
		seen[it.ItemID] = struct{}{}
		if it.Score < 0 {
			it.Score = 0
		}
		if it.Score > 1 {
			it.Score = 1
		}
		it.Reasons = cleanStrings(it.Reasons, 3, 20)
		it.Risk = cleanStrings(it.Risk, 3, 20)
		ranked = append(ranked, it)
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	if len(ranked) > 10 {
		ranked = ranked[:10]
	}
	r.Ranked = ranked

	r.Verdict = strings.ToLower(strings.TrimSpace(r.Verdict))
	if !IsValidVerdict(r.Verdict) {
		r.Verdict = deriveVerdict(ranked, strong, ambiguous)
	}
	// verdict 与得分自洽性校正：以得分为准，避免模型给出口径矛盾的结论
	if derived := deriveVerdict(ranked, strong, ambiguous); derived != "" && derived != r.Verdict && len(ranked) > 0 {
		r.Verdict = derived
	}
	if r.Verdict != VerdictAmbiguous {
		r.NeedMoreInfo = nil
	} else if r.NeedMoreInfo != nil {
		r.NeedMoreInfo.Question = truncateRunes(cleanStr(r.NeedMoreInfo.Question), 60)
		r.NeedMoreInfo.Purpose = truncateRunes(cleanStr(r.NeedMoreInfo.Purpose), 20)
		if r.NeedMoreInfo.Question == "" {
			r.NeedMoreInfo = nil
		}
	}
	r.Summary = truncateRunes(cleanStr(r.Summary), 60)
	// 条数由代码渲染：模型自报的数量不可靠，摘要中一旦出现数字或中文数词就直接丢弃整条摘要
	if containsCountLike(r.Summary) {
		r.Summary = ""
	}
	return &r, nil
}

// countLikeChars 数字与中文数词（含「两」「几」）：用于过滤模型自报的条数
const countLikeChars = "0123456789一二三四五六七八九十两百千万几"

// containsCountLike 判断文本是否包含数量类字符
func containsCountLike(s string) bool {
	return strings.ContainsAny(s, countLikeChars)
}

// deriveVerdict 按阈值从得分推导 verdict
func deriveVerdict(ranked []RerankItem, strong, ambiguous float64) string {
	if len(ranked) == 0 {
		return VerdictNoMatch
	}
	top := ranked[0].Score
	switch {
	case top >= strong:
		return VerdictStrong
	case top >= ambiguous:
		return VerdictAmbiguous
	default:
		return VerdictNoMatch
	}
}
