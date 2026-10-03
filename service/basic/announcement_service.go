package basic

import (
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type AnnouncementServiceGroup struct{}

const (
	announcementDefaultPage     = 1
	announcementDefaultPageSize = 10
	// 标题上限对齐 announcements.sql title varchar(100)，按字节判定：字节数 ≤ 100 ⇒ 字符数 ≤ 100，不会超出列宽
	announcementMaxTitleLen = 100

	// 公告类型（announcements.type）
	announcementTypeMin int8 = 0 // 0系统公告
	announcementTypeMax int8 = 3 // 3其他

	// 公告状态（announcements.status）：0 已废弃（历史垃圾行，不可见）
	announcementStatusPublished int8 = 1 // 已发布
	announcementStatusOffline   int8 = 2 // 已下架

	// 置顶标记（announcements.is_top）
	announcementIsTopMin int8 = 0 // 0否
	announcementIsTopMax int8 = 1 // 1是
)

// validateTitle 标题校验：trim 非空且 ≤100 字节（→ 90003）
func validateTitle(title string) response.Code {
	if strings.TrimSpace(title) == "" || len(title) > announcementMaxTitleLen {
		return response.CodeAnnouncementInvalid
	}
	return response.CodeSuccess
}

// validateContent 内容校验：trim 非空（→ 90003）
func validateContent(content string) response.Code {
	if strings.TrimSpace(content) == "" {
		return response.CodeAnnouncementInvalid
	}
	return response.CodeSuccess
}

// validateType 类型校验：0-3（→ 90003）
func validateType(t int8) response.Code {
	if t < announcementTypeMin || t > announcementTypeMax {
		return response.CodeAnnouncementInvalid
	}
	return response.CodeSuccess
}

// validateStatus 状态校验：仅 1已发布/2已下架（→ 90003）
func validateStatus(s int8) response.Code {
	if s != announcementStatusPublished && s != announcementStatusOffline {
		return response.CodeAnnouncementInvalid
	}
	return response.CodeSuccess
}

// validateIsTop 置顶校验：0/1（→ 90003）
func validateIsTop(t int8) response.Code {
	if t < announcementIsTopMin || t > announcementIsTopMax {
		return response.CodeAnnouncementInvalid
	}
	return response.CodeSuccess
}

// normalizeAnnouncementPage 归一化分页参数（binding 已拦截负数与 page_size>100，此处只兜底缺省/0）
func normalizeAnnouncementPage(q *model.AnnouncementListQuery) {
	if q.Page <= 0 {
		q.Page = announcementDefaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = announcementDefaultPageSize
	}
}

// Create 创建即发布：status=1、published_at=now、admin_id 取自 JWT（不信任请求体）
func (a *AnnouncementServiceGroup) Create(adminID int64, req *model.CreateAnnouncementRequest) response.Code {
	if code := validateTitle(req.Title); code != response.CodeSuccess {
		return code
	}
	if code := validateContent(req.Content); code != response.CodeSuccess {
		return code
	}
	if code := validateType(*req.Type); code != response.CodeSuccess {
		return code
	}
	if code := validateIsTop(*req.IsTop); code != response.CodeSuccess {
		return code
	}
	now := time.Now()
	announcement := &model.Announcement{
		AdminID:     adminID,
		Title:       req.Title,
		Content:     req.Content,
		Type:        *req.Type,
		Status:      announcementStatusPublished,
		IsTop:       *req.IsTop,
		PublishedAt: &now,
	}
	if err := dao.AnnouncementDao.CreateAnnouncement(announcement); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// Update 增量更新（仅传入字段；status 仅允许 1/2 下架或重新上架；published_at 保持首次发布时间不变）
func (a *AnnouncementServiceGroup) Update(req *model.AnnouncementUpdateRequest) response.Code {
	announcement, err := dao.AnnouncementDao.GetAnnouncementByID(req.ID)
	if err != nil {
		return response.CodeDatabaseError
	}
	if announcement == nil {
		return response.CodeAnnouncementNotFound
	}
	updates := make(map[string]interface{})
	if req.Title != nil {
		if code := validateTitle(*req.Title); code != response.CodeSuccess {
			return code
		}
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		if code := validateContent(*req.Content); code != response.CodeSuccess {
			return code
		}
		updates["content"] = *req.Content
	}
	if req.Type != nil {
		if code := validateType(*req.Type); code != response.CodeSuccess {
			return code
		}
		updates["type"] = *req.Type
	}
	if req.Status != nil {
		if code := validateStatus(*req.Status); code != response.CodeSuccess {
			return code
		}
		updates["status"] = *req.Status
	}
	if req.IsTop != nil {
		if code := validateIsTop(*req.IsTop); code != response.CodeSuccess {
			return code
		}
		updates["is_top"] = *req.IsTop
	}
	if err := dao.AnnouncementDao.UpdateAnnouncementByVK(req.ID, updates); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// Delete 软删除公告（is_deleted=1）
func (a *AnnouncementServiceGroup) Delete(id int64) response.Code {
	announcement, err := dao.AnnouncementDao.GetAnnouncementByID(id)
	if err != nil {
		return response.CodeDatabaseError
	}
	if announcement == nil {
		return response.CodeAnnouncementNotFound
	}
	if err := dao.AnnouncementDao.SoftDeleteAnnouncement(id); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// GetPublishedList 公开列表：仅 status=1（忽略客户端传入的 status），按 id 倒序，翻到底返回空列表
func (a *AnnouncementServiceGroup) GetPublishedList(q *model.AnnouncementListQuery) (*model.AnnouncementListResponse, response.Code) {
	return a.list(q, nil, true)
}

// GetAdminList 管理列表：含已下架（status 为 nil 时 1+2 全查），status 非 nil 按其筛选
func (a *AnnouncementServiceGroup) GetAdminList(q *model.AnnouncementListQuery) (*model.AnnouncementListResponse, response.Code) {
	return a.list(q, q.Status, false)
}

// list 分页列表公共实现
func (a *AnnouncementServiceGroup) list(q *model.AnnouncementListQuery, status *int8, publishedOnly bool) (*model.AnnouncementListResponse, response.Code) {
	normalizeAnnouncementPage(q)
	announcements, err := dao.AnnouncementDao.ListAnnouncements(q.Page, q.PageSize, status, publishedOnly)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	total, err := dao.AnnouncementDao.CountAnnouncements(status, publishedOnly)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	responses := make([]model.AnnouncementResponse, 0, len(announcements))
	for i := range announcements {
		responses = append(responses, *model.ToResponse(&announcements[i]))
	}
	return model.ToListResponse(q.Page, q.PageSize, total, responses), response.CodeSuccess
}

// GetDetail 公开详情：仅 status=1 可见（不存在/已下架/已删除 → 90001）；浏览量 +1，返回值含本次 +1
func (a *AnnouncementServiceGroup) GetDetail(id int64) (*model.AnnouncementResponse, response.Code) {
	announcement, err := dao.AnnouncementDao.GetAnnouncementByID(id)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	if announcement == nil || announcement.Status != announcementStatusPublished {
		return nil, response.CodeAnnouncementNotFound
	}
	// 浏览量写失败不影响详情返回（与 item 模块语义一致）
	_ = dao.AnnouncementDao.IncrViewCount(id)
	announcement.ViewCount++
	return model.ToResponse(announcement), response.CodeSuccess
}
