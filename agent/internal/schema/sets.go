package schema

// TagSet 三类标签词表（名称 → ID）；LLM 输出必须命中词表，否则被丢弃。
// 分组依据 tags.sort_order（见 model/mysql/data/insert_into_tags.sql）：
//
//	< 1000        物品类（含 999「其他」）
//	1000 ~ 1199   颜色
//	>= 1200       特征
type TagSet struct {
	Item    map[string]int64
	Color   map[string]int64
	Feature map[string]int64
}

// NewTagSet 构造空词表（非 nil，避免调用方判空）
func NewTagSet() TagSet {
	return TagSet{
		Item:    map[string]int64{},
		Color:   map[string]int64{},
		Feature: map[string]int64{},
	}
}

// ResolveItem 物品类标签名 → ID
func (t TagSet) ResolveItem(name string) (int64, bool) { id, ok := t.Item[name]; return id, ok }

// ResolveColor 颜色标签名 → ID
func (t TagSet) ResolveColor(name string) (int64, bool) { id, ok := t.Color[name]; return id, ok }

// ResolveFeature 特征标签名 → ID
func (t TagSet) ResolveFeature(name string) (int64, bool) { id, ok := t.Feature[name]; return id, ok }

// LocationSet 地点词表；LLM 只能输出 id，代码按 id 校验存在性
type LocationSet struct {
	ByID   map[int64]string   // id → 展示链路（如「屏峰校区/图书馆」，不含 L1 根节点）
	ByName map[string]int64   // 展示链路 → id
	ByLeaf map[string][]int64 // 末级名称 → id 列表（重名时为多项，如「1号」）
}

// NewLocationSet 构造空词表
func NewLocationSet() LocationSet {
	return LocationSet{
		ByID:   map[int64]string{},
		ByName: map[string]int64{},
		ByLeaf: map[string][]int64{},
	}
}

// Has 判断 id 是否存在于词表
func (l LocationSet) Has(id int64) bool {
	_, ok := l.ByID[id]
	return ok
}

// Name 返回展示链路
func (l LocationSet) Name(id int64) string { return l.ByID[id] }

// ResolveLeaf 按末级名称解析（唯一命中时返回；重名/不存在返回 false）
func (l LocationSet) ResolveLeaf(name string) (int64, bool) {
	ids, ok := l.ByLeaf[name]
	if !ok || len(ids) != 1 {
		return 0, false
	}
	return ids[0], true
}
