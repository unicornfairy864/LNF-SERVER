package vocab

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/agent/internal/schema"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

// ttl 词表缓存有效期：标签/地点由管理员低频维护，10 分钟足够新
const ttl = 10 * time.Minute

// tags.sort_order 分组边界（与 model/mysql/data/insert_into_tags.sql 的号段规则一致）
const (
	tagColorBoundary   = 1000 // < 1000：物品类（含 999「其他」）
	tagFeatureBoundary = 1200 // >= 1200：特征
)

// Vocabulary 运行时词表（tags + locations），供 prompt 注入与输出校验使用
type Vocabulary struct {
	mu       sync.RWMutex
	loadedAt time.Time

	tags       schema.TagSet
	locs       schema.LocationSet
	locLines   []string // 「30=屏峰校区/图书馆」，含 L2/L3/L4
	itemNames  []string // 物品类标签名（按 sort_order）
	colorNames []string // 颜色标签名
	featNames  []string // 特征标签名

	// 地点树辅助（检索用：全链搜索 + 忽略楼号）
	locParent   map[int64]int64
	locLevel    map[int64]int
	locChildren map[int64][]int64
	tagNameByID map[int64]string
}

var shared = &Vocabulary{}

// Shared 返回全局词表实例
func Shared() *Vocabulary { return shared }

// TagSet 返回标签词表（必要时触发加载/刷新）
func (v *Vocabulary) TagSet() (schema.TagSet, error) {
	if err := v.ensure(); err != nil {
		return schema.NewTagSet(), err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.tags, nil
}

// LocationSet 返回地点词表
func (v *Vocabulary) LocationSet() (schema.LocationSet, error) {
	if err := v.ensure(); err != nil {
		return schema.NewLocationSet(), err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.locs, nil
}

// LocationEnumLines 返回地点枚举行（格式 id=链路，按 level/id 升序；不含 L1 根节点）
func (v *Vocabulary) LocationEnumLines() ([]string, error) {
	if err := v.ensure(); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return copyStrings(v.locLines), nil
}

// GroupedTagNames 返回三类标签名（物品类 / 颜色 / 特征），供 prompt 分表注入
func (v *Vocabulary) GroupedTagNames() (item, color, feature []string, err error) {
	if err := v.ensure(); err != nil {
		return nil, nil, nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return copyStrings(v.itemNames), copyStrings(v.colorNames), copyStrings(v.featNames), nil
}

// Invalidate 主动失效缓存（词表变更后可由管理接口调用，当前未接入）
func (v *Vocabulary) Invalidate() {
	v.mu.Lock()
	v.loadedAt = time.Time{}
	v.mu.Unlock()
}

// NormalizeLocationForSearch 检索归一化：楼号（L4）上提到父级（L3），其余层级原样返回
// 依据用户定稿「查询忽略楼号」：西苑1号 与 西苑3号 在检索上等价于「西苑」
func (v *Vocabulary) NormalizeLocationForSearch(id int64) int64 {
	if err := v.ensure(); err != nil {
		return id
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.locLevel[id] == 4 {
		if p, ok := v.locParent[id]; ok && p > 0 {
			return p
		}
	}
	return id
}

// LocationScopeIDs 全链搜索集合：归一化后的地点 + 其全部祖先 + 全部后代（去重，含自身）
func (v *Vocabulary) LocationScopeIDs(ids []int64) []int64 {
	if err := v.ensure(); err != nil {
		return nil
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]int64, 0, len(ids)*4)
	seen := make(map[int64]struct{}, len(ids)*4)
	for _, raw := range ids {
		id := raw
		if v.locLevel[id] == 4 {
			if p, ok := v.locParent[id]; ok {
				id = p
			}
		}
		// 自身 + 祖先
		for cur := id; cur > 0; {
			if _, ok := seen[cur]; !ok {
				seen[cur] = struct{}{}
				out = append(out, cur)
			}
			next, ok := v.locParent[cur]
			if !ok || next <= 0 {
				break
			}
			cur = next
		}
		// 后代（BFS，限深防环）
		frontier := []int64{id}
		for depth := 0; depth < 5 && len(frontier) > 0; depth++ {
			var next []int64
			for _, f := range frontier {
				for _, c := range v.locChildren[f] {
					if _, ok := seen[c]; ok {
						continue
					}
					seen[c] = struct{}{}
					out = append(out, c)
					next = append(next, c)
				}
			}
			frontier = next
		}
	}
	return out
}

// TagNamesByIDs 按标签 ID 取名称（顺序与入参一致，未命中跳过；用于候选渲染）
func (v *Vocabulary) TagNamesByIDs(ids []int64) []string {
	if err := v.ensure(); err != nil {
		return nil
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := v.tagNameByID[id]; ok {
			out = append(out, name)
		}
	}
	return out
}

// ==================== 内部实现 ====================

// ensure 过期则重新加载；加载失败时若已有旧数据则降级使用旧数据
func (v *Vocabulary) ensure() error {
	v.mu.RLock()
	fresh := !v.loadedAt.IsZero() && time.Since(v.loadedAt) < ttl
	v.mu.RUnlock()
	if fresh {
		return nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.loadedAt.IsZero() && time.Since(v.loadedAt) < ttl {
		return nil
	}
	if err := v.loadLocked(); err != nil {
		if v.loadedAt.IsZero() {
			return err
		}
		return nil // 有旧数据：降级使用，不让 LLM 链路整体失败
	}
	return nil
}

// loadLocked 从数据库加载并构建词表（调用方需持有写锁）
func (v *Vocabulary) loadLocked() error {
	tags := dao.TagDao.GetTagList()
	locations := dao.LocationDao.GetLocations(nil, nil)
	if len(tags) == 0 || len(locations) == 0 {
		return fmt.Errorf("vocab: empty data (tags=%d locations=%d)", len(tags), len(locations))
	}

	tagSet := schema.NewTagSet()
	tagNameByID := make(map[int64]string, len(tags))
	itemNames := make([]string, 0, len(tags))
	colorNames := make([]string, 0, 16)
	featNames := make([]string, 0, 8)
	for _, t := range tags {
		tagNameByID[t.ID] = t.Name
		switch {
		case t.SortOrder < tagColorBoundary:
			tagSet.Item[t.Name] = t.ID
			itemNames = append(itemNames, t.Name)
		case t.SortOrder < tagFeatureBoundary:
			tagSet.Color[t.Name] = t.ID
			colorNames = append(colorNames, t.Name)
		default:
			tagSet.Feature[t.Name] = t.ID
			featNames = append(featNames, t.Name)
		}
	}

	locSet, locLines := buildLocations(locations)
	locParent := make(map[int64]int64, len(locations))
	locLevel := make(map[int64]int, len(locations))
	locChildren := make(map[int64][]int64, len(locations))
	for _, l := range locations {
		locParent[l.ID] = l.ParentID
		locLevel[l.ID] = l.Level
		if l.ParentID > 0 {
			locChildren[l.ParentID] = append(locChildren[l.ParentID], l.ID)
		}
	}

	v.tags = tagSet
	v.locs = locSet
	v.locLines = locLines
	v.itemNames = itemNames
	v.colorNames = colorNames
	v.featNames = featNames
	v.locParent = locParent
	v.locLevel = locLevel
	v.locChildren = locChildren
	v.tagNameByID = tagNameByID
	v.loadedAt = time.Now()
	return nil
}

// buildLocations 构建地点词表与枚举行（格式 id=链路，链路不含 L1 根节点）
func buildLocations(locations []modeladv.Location) (schema.LocationSet, []string) {
	byID := make(map[int64]modeladv.Location, len(locations))
	for _, l := range locations {
		byID[l.ID] = l
	}
	locSet := schema.NewLocationSet()
	lines := make([]string, 0, len(locations))
	// 排序：level 升序、同级按 id 升序（prompt 稳定，利于上游缓存与人工核对）
	sorted := make([]modeladv.Location, len(locations))
	copy(sorted, locations)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Level != sorted[j].Level {
			return sorted[i].Level < sorted[j].Level
		}
		return sorted[i].ID < sorted[j].ID
	})
	for _, l := range sorted {
		if l.Level <= 1 {
			continue // 根节点（浙江工业大学）不参与枚举
		}
		display := strings.Join(chainNames(l, byID), "/")
		if display == "" {
			continue
		}
		locSet.ByID[l.ID] = display
		locSet.ByName[display] = l.ID
		locSet.ByLeaf[l.Name] = append(locSet.ByLeaf[l.Name], l.ID)
		lines = append(lines, fmt.Sprintf("%d=%s", l.ID, display))
	}
	return locSet, lines
}

// chainNames 沿 parent 上溯取名称（返回 根→叶；跳过 L1 根节点；限深 10 防环）
func chainNames(l modeladv.Location, byID map[int64]modeladv.Location) []string {
	names := make([]string, 0, 4)
	cur := l
	for i := 0; i < 10; i++ {
		if cur.Level <= 1 {
			break
		}
		names = append(names, cur.Name)
		if cur.ParentID == 0 {
			break
		}
		next, ok := byID[cur.ParentID]
		if !ok {
			break
		}
		cur = next
	}
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return names
}

// copyStrings 复制字符串切片（避免调用方改动内部缓存）
func copyStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}
