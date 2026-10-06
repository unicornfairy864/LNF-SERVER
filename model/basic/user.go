package model

import "time"

// User 用户实体模型
type User struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_username" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Nickname     string     `gorm:"column:nickname;type:varchar(50);not null" json:"nickname"`
	Realname     *string    `gorm:"column:realname;type:varchar(50)" json:"realname,omitempty"`
	Gender       *int8      `gorm:"column:gender;type:tinyint;default:null" json:"gender,omitempty"`
	QQ           *string    `gorm:"column:qq;type:varchar(50);uniqueIndex:uk_qq" json:"qq,omitempty"`
	Avatar       *string    `gorm:"column:avatar;type:varchar(255)" json:"avatar,omitempty"`
	Role         int8       `gorm:"column:role;type:tinyint;not null;default:0" json:"role"`
	Status       int8       `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`
	Credit       int64      `gorm:"column:credit;type:int;not null;default:0" json:"credit"`
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
	Realname    *string    `json:"realname"`
	Gender      *int8      `json:"gender"`
	QQ          *string    `json:"qq"`
	Avatar      *string    `json:"avatar"`
	Role        int8       `json:"role"`
	Status      int8       `json:"status"`
	Credit      int64      `json:"credit"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PublicUserResponse 公开相应模型
type PublicUserResponse struct {
	ID          int64      `json:"id"`
	Nickname    string     `json:"nickname"`
	Gender      *int8      `json:"gender"`
	Avatar      *string    `json:"avatar"`
	Role        int8       `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LogoutRequest 用户登出请求
// logout_all 为指针：0（仅当前会话）是合法输入，用 required 校验"必须显式传入"
type LogoutRequest struct {
	LogoutAll *int8 `json:"logout_all" binding:"required"`
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
	QQ int64 `json:"qq" binding:"required"`
}

// QQBindRequest 绑定QQ请求
type QQBindRequest struct {
	QQ   int64 `json:"qq" binding:"required"`
	Code int64 `json:"code" binding:"required"`
}

// ChangeUserRoleRequest 用户角色变更请求
type ChangeUserRoleRequest struct {
	ID   int64 `json:"id" binding:"required"`
	Role *int8 `json:"role" binding:"required"`
}

// ChangeUserStatusRequest 用户状态变更请求
type ChangeUserStatusRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Status *int8 `json:"status" binding:"required"`
}

// AddUserCreditRequest 用户积分变更请求
// credit / type 为指针：0 是合法输入（credit=0 不产生变动，type=0 拾金不昧奖励）
type AddUserCreditRequest struct {
	ID         int64   `json:"id" binding:"required"`
	Credit     *int64  `json:"credit" binding:"required"`
	Type       *int64  `json:"type" binding:"required"`
	Desc       *string `json:"description,omitempty"`
	OperatorID int64   `json:"operator_id" binding:"required"`
}

type BatchRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// UserToResponse 将User模型转换为UserResponse
func UserToResponse(user *User) *UserResponse {
	if user == nil {
		return &UserResponse{}
	}

	return &UserResponse{
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

// UserToPublicUserResponse 将用户转为公开响应模型
func UserToPublicUserResponse(u *User) *PublicUserResponse {
	if u == nil {
		return nil
	}

	return &PublicUserResponse{
		ID:          u.ID,
		Nickname:    u.Nickname,
		Gender:      u.Gender,
		Avatar:      u.Avatar,
		Role:        u.Role,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

// CreditLog 积分变动流水模型（credit_logs）。
// 2026-10-07：为支撑「我的积分流水」查询补齐主键与列映射。此前该结构仅用于写入
// （无主键、无 TableName，靠 gorm 命名约定映射 credit_logs）；本次字段类型未变更，
// 写入路径（dao/user_dao.go AddUserCreditTx）行为不变。
// 字段说明：related_id 现有写入路径未填充（历史数据恒 NULL）；operator_id 列可空但写入路径恒赋 0；
// is_deleted 为内部逻辑删除位（查询恒过滤 =0，沿用项目惯例 json:"-" 不外泄）。
type CreditLog struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"column:user_id;type:bigint;not null;index:idx_user_created,priority:1" json:"user_id"`
	ChangeAmount int64     `gorm:"column:change_amount;type:int;not null" json:"change_amount"`
	BeforeAmount int64     `gorm:"column:before_amount;type:int;not null" json:"before_amount"`
	AfterAmount  int64     `gorm:"column:after_amount;type:int;not null" json:"after_amount"`
	Type         int64     `gorm:"column:type;type:tinyint;not null" json:"type"`
	RelatedID    *int64    `gorm:"column:related_id;type:bigint" json:"related_id,omitempty"`
	Description  string    `gorm:"column:description;type:varchar(255)" json:"description"`
	OperatorID   int64     `gorm:"column:operator_id;type:bigint" json:"operator_id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_user_created,priority:2" json:"created_at"`
	IsDeleted    int8      `gorm:"column:is_deleted;type:tinyint;not null;default:0" json:"-"`
}

// TableName 指定表名
func (CreditLog) TableName() string {
	return "credit_logs"
}

// CreditLogListQuery 我的积分流水查询条件（GET 参数）
// Type 为指针：0（拾金不昧奖励）是合法值，用指针区分「不筛」与「筛 0」；
// binding oneof 拦截越界值 → 参数错误；传空串（?type=）会因指针解析失败返回参数错误，不筛请省略参数
type CreditLogListQuery struct {
	Type     *int64 `form:"type,omitempty" binding:"omitempty,oneof=0 1 2 3 4"`
	Page     int    `form:"page,omitempty" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}

// CreditLogItem 积分流水列表项（对外字段白名单：不含 operator_id / related_id / is_deleted）
type CreditLogItem struct {
	ID           int64     `json:"id"`
	ChangeAmount int64     `json:"change_amount"`
	BeforeAmount int64     `json:"before_amount"`
	AfterAmount  int64     `json:"after_amount"`
	Type         int64     `json:"type"`
	TypeLabel    string    `json:"type_label"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreditLogListResponse 积分流水分页列表响应（空数据时 logs 为空数组而非 null）
type CreditLogListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Logs     []CreditLogItem `json:"logs"`
}
