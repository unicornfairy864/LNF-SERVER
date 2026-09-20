package model

import "time"

// User 用户实体模型
type User struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_username" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Nickname     string     `gorm:"column:nickname;type:varchar(50);not null" json:"nickname"`
	Realname     *string    `gorm:"column:rename;type:varchar(50)" json:"realname,omitempty"`
	Gender       *int8      `gorm:"column:gender;type:tinyint;default:null" json:"gender,omitempty"`
	QQ           *string    `gorm:"column:qq;type:varchar(50);uniqueIndex:uk_qq" json:"qq,omitempty"`
	Avatar       *string    `gorm:"column:avatar;type:varchar(255)" json:"avatar,omitempty"`
	Role         int8       `gorm:"column:role;type:tinyint;not null;default:0" json:"role"`
	Status       int8       `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`
	Credit       int        `gorm:"column:credit;type:int;not null;default:0" json:"credit"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at;type:datetime" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	IsDeleted    int8       `gorm:"column:is_deleted;type:tinyint;not null;default:0" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserResponse 用户响应模型
type UserResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Realname    *string    `json:"realname,omitempty"`
	Gender      *int8      `json:"gender,omitempty"`
	QQ          *string    `json:"qq,omitempty"`
	Avatar      *string    `json:"avatar,omitempty"`
	Role        int8       `json:"role"`
	Status      int8       `json:"status"`
	Credit      int        `json:"credit"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LogoutRequest 用户登录请求
type LogoutRequest struct {
	LogoutAll int8 `json:"logout_all" binding:"required"`
}

// CreateUserRequest 用户注册请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

// UpdateUserRequest 用户更新请求
type UpdateUserRequest struct {
	Nickname *string `json:"nickname,omitempty"`
	Realname *string `json:"realname,omitempty"`
	Gender   *int8   `json:"gender,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
}

// QQGetCodeRequest 获取qq验证码请求
type QQGetCodeRequest struct {
	QQ string `json:"qq"`
}

// QQBindRequest 绑定QQ请求
type QQBindRequest struct {
	QQ   string `json:"qq"`
	Code string `json:"code"`
}

// ChangeUserRoleRequest 用户角色变更请求
type ChangeUserRoleRequest struct {
	ID   int64 `json:"id" binding:"required"`
	Role int8  `json:"role" binding:"required"`
}

// ChangeUserStatusRequest 用户状态变更请求
type ChangeUserStatusRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Status int8  `json:"status" binding:"required"`
}

// ChangeUserCreditRequest 用户积分变更请求
type ChangeUserCreditRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Credit int   `json:"credit" binding:"required"`
}

// UserToResponse 将User模型转换为UserResponse
func UserToResponse(user *User) UserResponse {
	if user == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Realname:    user.Realname,
		Gender:      user.Gender,
		QQ:          user.QQ,
		Avatar:      user.Avatar,
		Role:        user.Role,
		Status:      user.Status,
		Credit:      user.Credit,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
