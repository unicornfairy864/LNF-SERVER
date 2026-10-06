package model

import "time"

// Comment 对应 comments 表（评论/反馈表）
type Comment struct {
	ID        int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement"                                     json:"id"`      // 主键ID
	ItemID    int64     `gorm:"column:item_id;type:bigint;not null;index:idx_item"                                 json:"item_id"` // 物品ID
	UserID    int64     `gorm:"column:user_id;type:bigint;not null;index:idx_user_created,priority:1"              json:"user_id"` // 评论用户ID
	RootID    int64     `gorm:"column:root_id;type:bigint"`
	ParentID  int64     `gorm:"column:parent_id;type:bigint;index:idx_parent_created,priority:1"                   json:"parent_id"`                                             // 父评论ID（支持回复评论）
	Content   string    `gorm:"column:content;type:text;not null"                                                  json:"content"`                                               // 评论内容（支持 markdown）
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:1"                                      json:"status"`                                                // 状态: 0待审核 1正常 2已隐藏
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime;index:idx_user_created,priority:2;index:idx_parent_created,priority:2" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime"                            json:"updated_at"`                                            // 更新时间
	IsDeleted int8      `gorm:"column:is_deleted;type:tinyint;not null;default:0"                                  json:"is_deleted"`                                            // 逻辑删除: 0否 1是
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// CommentDTO 用于 API 响应，隐藏 is_deleted 等内部字段
type CommentDTO struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	UserID    int64     `json:"user_id"`
	ParentID  int64     `json:"parent_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDTO 将数据库实体转换为 API 响应对象
func (c *Comment) ToDTO() *CommentDTO {
	if c == nil {
		return nil
	}
	return &CommentDTO{
		ID:        c.ID,
		ItemID:    c.ItemID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type CommentsListDTO struct {
	Total       int           `json:"total"`
	CommentDTOs []*CommentDTO `json:"comment_dtos"`
}

func ToList(total int, cDTOs []*CommentDTO) *CommentsListDTO {
	if total == len(cDTOs) {
		return &CommentsListDTO{
			Total:       total,
			CommentDTOs: cDTOs,
		}
	}
	return nil
}

// CreateCommentRequest 创建评论的请求体
type CreateCommentRequest struct {
	ItemID   int64  `json:"item_id" binding:"required"`
	UserID   int64  `json:"user_id" binding:"required"`
	ParentID *int64 `json:"parent_id"`
	Content  string `json:"content" binding:"required"`
}

// ToModel 将创建请求转换为数据库实体（需传入当前用户 ID）
func (r *CreateCommentRequest) ToModel(userID int64) *Comment {
	if r == nil {
		return nil
	}
	return &Comment{
		ItemID:   r.ItemID,
		UserID:   userID,
		ParentID: *r.ParentID,
		Content:  r.Content,
	}
}

// UpdateCommentRequest 更新评论的请求体
type UpdateCommentRequest struct {
	ID     *int64 `json:"id" binding:"required"`
	Status *int8  `json:"status" binding:"required"` //仅管理员或特定场景使用
}

//获取评论列表的请求体
type CommentGetlistRequestQuery struct {
	StartedID int64 `form:"started_id"`
	Limit     int64 `form:"limit"`
}
type CommentGetListRequestParam struct {
	ItemID int64 `uri:"itemID"`
}

//获取子节点的请求体
type GetChildrenRequest struct {
	ID       int64 `json:"id" form:"id" binding:"required"`
	Depth    int64 `json:"depth" form:"depth"`
	MaxCount int64 `json:"max_count" form:"max_count"` // 总节点数上限
}

//dao获取comments的入参结构体
type CommentsQuery struct {
	ID        *int64
	ItemID    *int64
	RootID    *int64
	UserID    *int64
	StartedID *int64
	Limit     int64
	IDs       *[]int64
}
