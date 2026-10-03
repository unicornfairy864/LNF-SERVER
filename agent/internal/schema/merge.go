package schema

import "strings"

// ==================== 建帖确认轮的「补充信息」合并结果 ====================

const (
	DecisionConfirm     = "confirm"      // 用户确认发布
	DecisionCancel      = "cancel"       // 用户放弃发布
	DecisionProvideInfo = "provide_info" // 用户补充/修改信息（带 patch）
	DecisionUnrelated   = "unrelated"    // 其他无关内容
)

// IsValidDecision 判断 decision 是否合法
func IsValidDecision(v string) bool {
	switch v {
	case DecisionConfirm, DecisionCancel, DecisionProvideInfo, DecisionUnrelated:
		return true
	}
	return false
}

// MergePatch 只包含需要更新的字段（未提及的字段不出现在 patch 中 = 增量语义）
type MergePatch struct {
	Title          *string   `json:"title"`
	Description    *string   `json:"description"`
	ItemTag        *string   `json:"item_tag"`
	ItemTagID      *int64    `json:"item_tag_id,omitempty"`
	ColorTag       *string   `json:"color_tag"`
	ColorTagID     *int64    `json:"color_tag_id,omitempty"`
	FeatureTags    *[]string `json:"feature_tags"`
	FeatureTagIDs  *[]int64  `json:"feature_tag_ids,omitempty"`
	LocationID     *int64    `json:"location_id"`
	LocationDetail *string   `json:"location_detail"`
	Contact        *string   `json:"contact"` // 仅当用户明确给出联系方式时填写（不编造）
	TimeFrom       *string   `json:"time_from"`
	TimeTo         *string   `json:"time_to"`
}

// MergeResult 合并判定结果
type MergeResult struct {
	Decision string      `json:"decision"`
	Patch    *MergePatch `json:"patch"`
	Note     string      `json:"note"`
}

// ValidateMerge 解析并校验「补充信息合并」输出：
//   - decision 非法 → unrelated
//   - 仅 provide_info 保留 patch；其余情况 patch 一律丢弃
//   - patch 内标签/地点必须命中词表；时间非法置空；长度对齐 SQL 列宽
func ValidateMerge(raw string, tags TagSet, locs LocationSet) (*MergeResult, error) {
	var r MergeResult
	if err := UnmarshalObject(raw, &r); err != nil {
		return nil, err
	}
	r.Decision = strings.ToLower(strings.TrimSpace(r.Decision))
	if !IsValidDecision(r.Decision) {
		r.Decision = DecisionUnrelated
	}
	r.Note = truncateRunes(cleanStr(r.Note), 60)
	if r.Decision != DecisionProvideInfo {
		r.Patch = nil
		return &r, nil
	}
	if r.Patch == nil {
		r.Decision = DecisionUnrelated
		return &r, nil
	}
	p := r.Patch
	if p.Title != nil {
		t := truncateBytes(cleanStr(*p.Title), 100)
		p.Title = cleanStrPtr(t)
	}
	if p.Description != nil {
		d := truncateRunes(cleanStr(*p.Description), 200)
		p.Description = cleanStrPtr(d)
	}
	p.ItemTag, p.ItemTagID = resolveTag(p.ItemTag, tags.ResolveItem)
	p.ColorTag, p.ColorTagID = resolveTag(p.ColorTag, tags.ResolveColor)
	if p.FeatureTags != nil {
		names, ids := resolveTags(*p.FeatureTags, tags.ResolveFeature)
		if names == nil {
			p.FeatureTags, p.FeatureTagIDs = nil, nil
		} else {
			p.FeatureTags, p.FeatureTagIDs = &names, &ids
		}
	}
	if p.LocationID != nil && !locs.Has(*p.LocationID) {
		p.LocationID = nil
	}
	if p.LocationDetail != nil {
		p.LocationDetail = cleanStrPtr(truncateRunes(cleanStr(*p.LocationDetail), 200))
	}
	if p.Contact != nil {
		p.Contact = cleanStrPtr(truncateBytes(cleanStr(*p.Contact), 100))
	}
	if parseRFC3339(p.TimeFrom) == nil {
		p.TimeFrom = nil
	}
	if parseRFC3339(p.TimeTo) == nil {
		p.TimeTo = nil
	}
	// patch 全空 → 视为无关
	if p.Title == nil && p.Description == nil && p.ItemTag == nil && p.ColorTag == nil &&
		p.FeatureTags == nil && p.LocationID == nil && p.LocationDetail == nil &&
		p.Contact == nil && p.TimeFrom == nil && p.TimeTo == nil {
		r.Decision = DecisionUnrelated
		r.Patch = nil
	}
	return &r, nil
}
