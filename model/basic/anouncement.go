package model

import "time"

type Announcement struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"                   json:"id"           comment:"主键ID"`
	AdminID     int64      `gorm:"column:admin_id;not null"                             json:"admin_id"     comment:"发布管理员ID"`
	Title       string     `gorm:"column:title;type:varchar(100);not null"              json:"title"        comment:"公告标题"`
	Content     string     `gorm:"column:content;type:text;not null"                    json:"content"      comment:"公告内容（支持markdown格式）"`
	Type        int8       `gorm:"column:type;not null"                                 json:"type"         comment:"公告类型: 0系统公告 1活动公告 2维护通知 3其他"`
	Status      int8       `gorm:"column:status;not null;default:0"                     json:"status"       comment:"状态: 0草稿 1已发布 2已过期"`
	IsTop       int8       `gorm:"column:is_top;not null;default:0"                     json:"is_top"       comment:"是否置顶: 0否 1是"`
	ViewCount   int32      `gorm:"column:view_count;not null;default:0"                 json:"view_count"   comment:"浏览次数"`
	PublishedAt *time.Time `gorm:"column:published_at"                                  json:"published_at" comment:"发布时间"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"   comment:"创建时间"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"   comment:"更新时间"`
	IsDeleted   int8       `gorm:"column:is_deleted;not null;default:0"                 json:"is_deleted"   comment:"逻辑删除: 0否 1是"`
}

//创建更新都用这个
type AnnouncementUpdateRequest struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"                   json:"id"           comment:"主键ID"`
	AdminID   int64  `gorm:"column:admin_id;not null"                             json:"admin_id"     comment:"发布管理员ID"`
	Title     string `gorm:"column:title;type:varchar(100);not null"              json:"title"        comment:"公告标题"`
	Content   string `gorm:"column:content;type:text;not null"                    json:"content"      comment:"公告内容（支持markdown格式）"`
	Type      int8   `gorm:"column:type;not null"                                 json:"type"         comment:"公告类型: 0系统公告 1活动公告 2维护通知 3其他"`
	Status    int8   `gorm:"column:status;not null;default:0"                     json:"status"       comment:"状态: 0草稿 1已发布 2已过期"`
	IsTop     int8   `gorm:"column:is_top;not null;default:0"                     json:"is_top"       comment:"是否置顶: 0否 1是"`
	IsDeleted int8   `gorm:"column:is_deleted;not null;default:0"                 json:"is_deleted"   comment:"逻辑删除: 0否 1是"`
}

type AnnouncementGetRequest struct {
	Auth         bool  `json:"auth"`
	AdminID      int64 `json:"admin_id"`
	StartedID    int64 `json:"started_id"`
	IgnorePieces int64 `json:"ignore_pieces"`
	Limit        int64 `json:"limit"`
}

type AnnouncementResponse struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"                   json:"id"           comment:"主键ID"`
	AdminID     int64      `gorm:"column:admin_id;not null"                             json:"admin_id"     comment:"发布管理员ID"`
	Title       string     `gorm:"column:title;type:varchar(100);not null"              json:"title"        comment:"公告标题"`
	Content     string     `gorm:"column:content;type:text;not null"                    json:"content"      comment:"公告内容（支持markdown格式）"`
	Type        int8       `gorm:"column:type;not null"                                 json:"type"         comment:"公告类型: 0系统公告 1活动公告 2维护通知 3其他"`
	Status      int8       `gorm:"column:status;not null;default:0"                     json:"status"       comment:"状态: 0草稿 1已发布 2已过期"`
	IsTop       int8       `gorm:"column:is_top;not null;default:0"                     json:"is_top"       comment:"是否置顶: 0否 1是"`
	ViewCount   int32      `gorm:"column:view_count;not null;default:0"                 json:"view_count"   comment:"浏览次数"`
	PublishedAt *time.Time `gorm:"column:published_at"                                  json:"published_at" comment:"发布时间"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"   comment:"创建时间"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"   comment:"更新时间"`
}

type AnnouncementListResponse struct {
	Total         int64                  `json:"total"`
	Announcements []AnnouncementResponse `json:"announcements"`
}

func ToAnnouncement(r *AnnouncementUpdateRequest) *Announcement {
	if r == nil {
		return nil
	}
	return &Announcement{
		ID:        r.ID,
		AdminID:   r.AdminID,
		Title:     r.Title,
		Content:   r.Content,
		Type:      r.Type,
		Status:    r.Status,
		IsTop:     r.IsTop,
		IsDeleted: r.IsDeleted,
	}
}

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

func ToListResponse(total int64, announcements []AnnouncementResponse) *AnnouncementListResponse {
	return &AnnouncementListResponse{
		Total:         total,
		Announcements: announcements,
	}
}
