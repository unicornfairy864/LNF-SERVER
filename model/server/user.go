package server

import "time"

type UserResponse struct {
    Username      string    `json:"username"`
    RealName      string    `json:"real_name"`
    QQ            string    `json:"qq"`
    Avatar        string    `json:"avatar"`
    Role          int8      `json:"role"`
    Credit        int       `json:"credit"`
    LastLoginTime time.Time `json:"last_login_time"`
    CreateTime    time.Time `json:"create_time"`
}

type User struct {
    ID            int64     `json:"id"`
    Username      string    `json:"username"`
    PasswordHash  string    `json:"password"`
    RealName      string    `json:"real_name"`
    QQ            string    `json:"qq"`
    Avatar        string    `json:"avatar"`
    Role          int8      `json:"role"`
    Status        int8      `json:"status"`
    Credit        int       `json:"credit"`
    LastLoginTime time.Time `json:"last_login_time"`
    CreateTime    time.Time `json:"create_time"`
    UpdateTime    time.Time `json:"update_time"`
    IsDeleted     int8      `json:"is_deleted"`
}

func UserToResponse(u User) UserResponse {
    return UserResponse{
        Username:      u.Username,
        RealName:      u.RealName,
        QQ:            u.QQ,
        Avatar:        u.Avatar,
        Role:          u.Role,
        Credit:        u.Credit,
        LastLoginTime: u.LastLoginTime,
        CreateTime:    u.CreateTime,
    }
}