package schema

import (
	"strings"
	"time"
)

// ==================== 意图 / 语境枚举 ====================

const (
	IntentCreateLost  = "create_lost"  // 用户丢失物品，求线索
	IntentCreateFound = "create_found" // 用户捡到物品，招领
	IntentMatch       = "match"        // 用户想匹配相似帖子
	IntentChitchat    = "chitchat"     // 闲聊（与失物招领无关）
	IntentOther       = "other"        // 无法归类 / 信息不足 / 非现实语境
)

// IsValidIntent 判断意图是否合法
func IsValidIntent(v string) bool {
	switch v {
	case IntentCreateLost, IntentCreateFound, IntentMatch, IntentChitchat, IntentOther:
		return true
	}
	return false
}

// IsCreateIntent 是否为「建帖」类意图
func IsCreateIntent(v string) bool { return v == IntentCreateLost || v == IntentCreateFound }

// missing_fields 允许值（提示前端/QQ 侧需要追问什么）
var allowedMissingFields = map[string]struct{}{
	"location": {}, "location_detail": {}, "time": {}, "contact": {}, "color": {}, "features": {},
}

// ==================== 抽取结构 ====================

// ExtractItem 单个物品的抽取结果
// TimeFrom/TimeTo 保留模型原始字符串（RFC3339），校验失败时置 nil；用 Times() 取解析值。
type ExtractItem struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ItemTag        *string  `json:"item_tag"`
	ItemTagID      *int64   `json:"item_tag_id,omitempty"`
	ColorTag       *string  `json:"color_tag"`
	ColorTagID     *int64   `json:"color_tag_id,omitempty"`
	FeatureTags    []string `json:"feature_tags"`
	FeatureTagIDs  []int64  `json:"feature_tag_ids,omitempty"`
	Keywords       []string `json:"keywords"`
	LocationID     *int64   `json:"location_id"`
	LocationDetail *string  `json:"location_detail"`
	TimeFrom       *string  `json:"time_from"`
	TimeTo         *string  `json:"time_to"`
	Confidence     float64  `json:"confidence"`
}

// Times 解析时间区间（东八区 RFC3339）；任一端非法则返回 nil
func (i *ExtractItem) Times() (from, to *time.Time) {
	from = parseRFC3339(i.TimeFrom)
	to = parseRFC3339(i.TimeTo)
	if from != nil && to != nil && from.After(*to) {
		from, to = to, from
	}
	return from, to
}

// ExtractResult 判类 + 抽取的完整输出
type ExtractResult struct {
	Intent           string        `json:"intent"`
	IsLNFContext     bool          `json:"is_lnf_context"`
	Items            []ExtractItem `json:"items"`
	MissingFields    []string      `json:"missing_fields"`
	FollowupQuestion string        `json:"followup_question"`
}

// FirstItem 返回首个物品（无则 nil）
func (r *ExtractResult) FirstItem() *ExtractItem {
	if len(r.Items) == 0 {
		return nil
	}
	return &r.Items[0]
}

// ValidateExtract 解析 + 白名单校验模型输出：
//   - 未知字段自动忽略（一次性解码到结构体）
//   - 意图非法 → other；标签不在词表 → 置空；地点不在词表 → 置空
//   - title ≤100 字节、description ≤200 字、keywords ≤6 个、missing_fields 仅允许枚举值
//   - 时间非法 → 置空（不整条失败，避免一条坏字段丢掉整个结果）
func ValidateExtract(raw string, tags TagSet, locs LocationSet) (*ExtractResult, error) {
	var r ExtractResult
	if err := UnmarshalObject(raw, &r); err != nil {
		return nil, err
	}
	r.Intent = strings.ToLower(strings.TrimSpace(r.Intent))
	if !IsValidIntent(r.Intent) {
		r.Intent = IntentOther
	}
	if r.Intent == IntentChitchat || r.Intent == IntentOther {
		r.IsLNFContext = false
	}
	// 非失物招领语境一律降级为 other，防止模型自相矛盾
	if !r.IsLNFContext {
		r.Intent = IntentOther
	}

	items := make([]ExtractItem, 0, len(r.Items))
	for idx := range r.Items {
		it := r.Items[idx]
		it.Title = truncateBytes(cleanStr(it.Title), 100)
		it.Description = truncateRunes(cleanStr(it.Description), 200)
		// 标签：必须命中词表（各分组独立），未命中即置空
		it.ItemTag, it.ItemTagID = resolveTag(it.ItemTag, tags.ResolveItem)
		it.ColorTag, it.ColorTagID = resolveTag(it.ColorTag, tags.ResolveColor)
		it.FeatureTags, it.FeatureTagIDs = resolveTags(it.FeatureTags, tags.ResolveFeature)
		it.Keywords = cleanStrings(it.Keywords, 6, 20)
		// 地点：必须命中词表
		if it.LocationID != nil && !locs.Has(*it.LocationID) {
			it.LocationID = nil
		}
		it.LocationDetail = cleanStrPtr(truncateRunes(ptrString(it.LocationDetail), 200))
		// 时间：非法则置空
		if parseRFC3339(it.TimeFrom) == nil {
			it.TimeFrom = nil
		}
		if parseRFC3339(it.TimeTo) == nil {
			it.TimeTo = nil
		}
		// 置信度收敛到 [0,1]
		if it.Confidence < 0 {
			it.Confidence = 0
		}
		if it.Confidence > 1 {
			it.Confidence = 1
		}
		if it.Title == "" && it.Description == "" {
			continue // 空条目直接丢弃
		}
		items = append(items, it)
	}
	r.Items = items

	// missing_fields：仅保留枚举内的值
	mf := make([]string, 0, len(r.MissingFields))
	seen := make(map[string]struct{}, len(r.MissingFields))
	for _, f := range r.MissingFields {
		f = strings.ToLower(strings.TrimSpace(f))
		if _, ok := allowedMissingFields[f]; !ok {
			continue
		}
		if _, dup := seen[f]; dup {
			continue
		}
		seen[f] = struct{}{}
		mf = append(mf, f)
		if len(mf) >= 6 {
			break
		}
	}
	r.MissingFields = mf
	r.FollowupQuestion = truncateRunes(cleanStr(r.FollowupQuestion), 100)
	if len(r.Items) == 0 && IsCreateIntent(r.Intent) {
		r.MissingFields = append(r.MissingFields, "item")
	}
	return &r, nil
}

// ==================== 内部工具 ====================

// resolveTag 单标签校验：命中词表则同时返回名称与 ID，否则双空
func resolveTag(name *string, resolve func(string) (int64, bool)) (*string, *int64) {
	if name == nil {
		return nil, nil
	}
	n := cleanStr(*name)
	if n == "" {
		return nil, nil
	}
	id, ok := resolve(n)
	if !ok {
		return nil, nil
	}
	return &n, &id
}

// resolveTags 多标签校验：仅保留命中词表的项
func resolveTags(names []string, resolve func(string) (int64, bool)) ([]string, []int64) {
	if len(names) == 0 {
		return nil, nil
	}
	outNames := make([]string, 0, len(names))
	outIDs := make([]int64, 0, len(names))
	seen := make(map[int64]struct{}, len(names))
	for _, raw := range names {
		n := cleanStr(raw)
		if n == "" {
			continue
		}
		id, ok := resolve(n)
		if !ok {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		outNames = append(outNames, n)
		outIDs = append(outIDs, id)
		if len(outNames) >= 4 {
			break
		}
	}
	if len(outNames) == 0 {
		return nil, nil
	}
	return outNames, outIDs
}

// parseRFC3339 解析 RFC3339（含时区）；非法返回 nil
func parseRFC3339(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &t
}

// ptrString 安全解引用
func ptrString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
