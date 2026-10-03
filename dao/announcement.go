package dao

import (
	"errors"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"gorm.io/gorm"
)

type AnnouncementGroup struct{}

// buildAnnouncementQuery 组装公告查询条件（每次调用返回全新 DB，避免 Count 污染链）：
// 固定 is_deleted=0；publishedOnly=true 强制 status=1（公开端，忽略 status 参数）；
// 否则 status 非 nil 按其筛选；nil 时 status IN (1,2)（管理端：1已发布+2已下架，0 废弃垃圾行不可见）
func buildAnnouncementQuery(status *int8, publishedOnly bool) *gorm.DB {
	db := global.LNF_DB.Model(&model.Announcement{}).Where("is_deleted = 0")
	switch {
	case publishedOnly:
		db = db.Where("status = ?", 1)
	case status != nil:
		db = db.Where("status = ?", *status)
	default:
		db = db.Where("status IN ?", []int8{1, 2})
	}
	return db
}

// ListAnnouncements 分页查询公告（单条 SQL，ORDER BY id DESC + Offset/Limit）
func (a *AnnouncementGroup) ListAnnouncements(page int, pageSize int, status *int8, publishedOnly bool) ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := buildAnnouncementQuery(status, publishedOnly).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&announcements).Error
	return announcements, err
}

// CountAnnouncements 统计满足条件的公告总数（与 ListAnnouncements 同口径）
func (a *AnnouncementGroup) CountAnnouncements(status *int8, publishedOnly bool) (int64, error) {
	var count int64
	err := buildAnnouncementQuery(status, publishedOnly).Count(&count).Error
	return count, err
}

// GetAnnouncementByID 按 ID 查询未删除公告；记录不存在返回 (nil, nil)，其他错误原样返回
func (a *AnnouncementGroup) GetAnnouncementByID(id int64) (*model.Announcement, error) {
	var announcement model.Announcement
	err := global.LNF_DB.Where(ConditionIDNotDeleted, id).First(&announcement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &announcement, nil
}

// CreateAnnouncement 创建公告（status/published_at/admin_id 由 service 构造：创建即发布）
func (a *AnnouncementGroup) CreateAnnouncement(announcement *model.Announcement) error {
	return global.LNF_DB.Create(announcement).Error
}

// UpdateAnnouncementByVK 按 ID 增量更新公告（白名单字段，带 is_deleted=0 守卫）
// 无有效字段时直接成功（不产生空 UPDATE）
func (a *AnnouncementGroup) UpdateAnnouncementByVK(id int64, updates map[string]interface{}) error {
	allowed := map[string]struct{}{
		"title": {}, "content": {}, "type": {}, "status": {}, "is_top": {},
	}
	filtered := make(map[string]interface{}, len(updates)+1)
	for k, v := range updates {
		if _, ok := allowed[k]; ok {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	filtered["updated_at"] = time.Now()
	return global.LNF_DB.Model(&model.Announcement{}).
		Where(ConditionIDNotDeleted, id).
		Updates(filtered).Error
}

// SoftDeleteAnnouncement 软删除公告（is_deleted=1，替换硬删）
func (a *AnnouncementGroup) SoftDeleteAnnouncement(id int64) error {
	return global.LNF_DB.Model(&model.Announcement{}).
		Where(ConditionIDNotDeleted, id).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"updated_at": time.Now(),
		}).Error
}

// IncrViewCount 浏览量 +1（仅 status=1 且未删除的公告）
func (a *AnnouncementGroup) IncrViewCount(id int64) error {
	return global.LNF_DB.Model(&model.Announcement{}).
		Where("id = ? AND is_deleted = 0 AND status = 1", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
