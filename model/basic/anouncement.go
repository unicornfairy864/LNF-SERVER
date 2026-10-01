package model

import "time"

// Announcement 公告实体模型（对应 announcements 表）
type Announcement struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"                   json:"id"                     comment:"主键ID"`
	AdminID     int64      `gorm:"column:admin_id;not null"                             json:"admin_id"               comment:"发布管理员ID"`
	Title       string     `gorm:"column:title;type:varchar(100);not null"              json:"title"                  comment:"公告标题"`
	Content     string     `gorm:"column:content;type:text;not null"                    json:"content"                comment:"公告内容（支持markdown格式）"`
	Type        int8       `gorm:"column:type;not null"                                 json:"type"                   comment:"公告类型: 0系统公告 1活动公告 2维护通知 3其他"`
	Status      int8       `gorm:"column:status;not null;default:0"                     json:"status"                 comment:"状态: 1已发布 2已下架（0已废弃）"`
	IsTop       int8       `gorm:"column:is_top;not null;default:0"                     json:"is_top"                 comment:"是否置顶: 0否 1是"`
	ViewCount   int32      `gorm:"column:view_count;not null;default:0"                 json:"view_count"             comment:"浏览次数"`
	PublishedAt *time.Time `gorm:"column:published_at"                                  json:"published_at,omitempty" comment:"发布时间"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"             comment:"创建时间"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"             comment:"更新时间"`
	IsDeleted   int8       `gorm:"column:is_deleted;not null;default:0"                 json:"-"                     comment:"逻辑删除: 0否 1是"`
}

// AnnouncementListQuery 公告列表查询条件（GET query）
// page/page_size 缺省（未传或为 0）由 service 归一化为 1/10；负数与 page_size>100 由 binding 拦截 → 1
// status 仅管理端生效（公开端强制 status=1），可选 0/1/2：0 是合法值（查历史废弃行），bind 不报错
type AnnouncementListQuery struct {
	Page     int   `form:"page,omitempty" binding:"omitempty,min=1"`
	PageSize int   `form:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
	Status   *int8 `form:"status,omitempty" binding:"omitempty,oneof=0 1 2"`
}

// CreateAnnouncementRequest 创建公告请求（创建即发布：status=1、published_at=now、admin_id 取自 JWT）
// 缺字段/JSON 类型错 → 绑定失败 → 1；字段业务校验（trim 非空、长度、取值范围）→ 90003
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    *int8  `json:"type" binding:"required"`
	IsTop   *int8  `json:"is_top" binding:"required"`
}

// AnnouncementUpdateRequest 公告增量更新请求：仅更新传入的字段（指针语义）
// 软删只走 delete 端点，客户端不可控 is_deleted；published_at/created_at/view_count 不可更新
type AnnouncementUpdateRequest struct {
	ID      int64   `json:"id" binding:"required"`
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Type    *int8   `json:"type,omitempty"`
	Status  *int8   `json:"status,omitempty"`
	IsTop   *int8   `json:"is_top,omitempty"`
}

// AnnouncementResponse 公告响应模型（无 is_deleted）
type AnnouncementResponse struct {
	ID          int64      `json:"id"`
	AdminID     int64      `json:"admin_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Type        int8       `json:"type"`
	Status      int8       `json:"status"`
	IsTop       int8       `json:"is_top"`
	ViewCount   int32      `json:"view_count"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AnnouncementListResponse 公告分页列表响应
type AnnouncementListResponse struct {
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	Announcements []AnnouncementResponse `json:"announcements"`
}

// ToResponse 实体转响应模型
func ToResponse(a *Announcement) *AnnouncementResponse {
	if a == nil {
		return nil
	}
	return &AnnouncementResponse{
		ID:          a.ID,
		AdminID:     a.AdminID,
		Title:       a.Title,
		Content:     a.Content,
		Type:        a.Type,
		Status:      a.Status,
		IsTop:       a.IsTop,
		ViewCount:   a.ViewCount,
		PublishedAt: a.PublishedAt,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// ToListResponse 组装分页列表响应
func ToListResponse(page int, pageSize int, total int64, announcements []AnnouncementResponse) *AnnouncementListResponse {
	return &AnnouncementListResponse{
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		Announcements: announcements,
	}
}
