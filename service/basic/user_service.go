package basic

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/unicornfairy864/LNF-SERVER/chensong"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type UserServiceGroup struct{}

func (userService *UserServiceGroup) Create(req *model.CreateUserRequest) (*model.User, response.Code) {
	// 判断表单是否符合要求
	if req.Username == "" || req.Nickname == "" || req.Password == "" ||
		len(req.Username) < 2 || len(req.Username) > 32 ||
		len(req.Nickname) < 2 || len(req.Nickname) > 32 ||
		len(req.Password) < 8 || len(req.Password) > 20 {
		return nil, response.CodeParamError
	}
	// 用户名是否被占用
	if dao.UserDao.GetUserByUsername(req.Username).ID != 0 {
		return nil, response.CodeUsernameOccupied
	}
	user, err := dao.UserDao.CreateUser(&model.User{
		Username:     req.Username,
		PasswordHash: utils.Bycrypt.GeneratePasswordHash(req.Password),
		Nickname:     req.Nickname,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return nil, response.CodeServerError
	}
	return &user, response.CodeSuccess
}

func (userService *UserServiceGroup) Login(req *model.LoginRequest) (*model.User, *string, response.Code) {
	// 判断表单是否符合要求
	if req.Username == "" || req.Password == "" ||
		len(req.Username) < 2 || len(req.Username) > 32 ||
		len(req.Password) < 8 || len(req.Password) > 20 {
		return nil, nil, response.CodeFormInvalid
	}
	// 用户是否存在
	user := dao.UserDao.GetUserByUsername(req.Username)
	if user.ID == 0 {
		return nil, nil, response.CodeUserOrPasswordError
	}
	if !utils.Bycrypt.CheckPassword(req.Password, user.PasswordHash) {
		return nil, nil, response.CodeUserOrPasswordError
	}
	if user.IsDeleted != 0 {
		return nil, nil, response.CodeUserNotFoundOrBanned
	}
	// 生成 JWTToken
	token, err := utils.JWT.GenerateToken(user, false)
	if err != nil {
		return nil, nil, response.CodeServerError
	}
	return user, &token, response.CodeSuccess
}

func (userService *UserServiceGroup) Logout(jti string, expiredAt time.Time) response.Code {
	utils.LogJson(jti)
	utils.LogJson(expiredAt)
	err := utils.JWT.BanToken(&utils.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	})
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) LogoutAll(uid int64) response.Code {
	err := utils.JWT.BanUserById(uid)
	if err != nil && !errors.Is(err, fmt.Errorf("VersionNotExist")) {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) Update(id int64, req *model.UpdateUserRequest) response.Code {
	updates := make(map[string]interface{}, 5)
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Realname != nil {
		updates["realname"] = *req.Realname
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if len(updates) == 0 {
		return response.CodeSuccess
	}
	updates["updated_at"] = time.Now()
	result := global.LNF_DB.Model(&model.User{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates)
	if result.Error != nil {
		return response.CodeDatabaseError
	}
	if result.RowsAffected == 0 {
		return response.CodeUserNotFoundOrBanned
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) QQGetCode(qq string, nickname string, jti string) response.Code {
	/* 绑定QQ思路: 若上一个 code 未失效, 则不生成新的
	 * 1, qq:bind:{jti}:code  string
	 * 2, qq:bind:{jti}:tries int64    <-Redis Incr Returns A Int64 Number
	 */
	// 判断会话是否存在
	str, err := dao.RedisDao.GetValueString("qq:bind:" + jti + ":code")
	if err != nil {
		return response.CodeDatabaseError
	}
	if str != "" {
		return response.CodeQQCodeAlreadyExists
	}
	ok, err := dao.RedisDao.KeyExists("qq:bind-session:" + qq)
	if err != nil || ok == true {
		return response.CodeDatabaseError
	}
	// 先生成验证码，对jti进行冷却，防止多次bot查询
	QQCode := strconv.Itoa(rand.Intn(900000) + 100000)
	kv := make(map[string]interface{})
	kv["qq:bind-session:"+qq] = 1
	kv["qq:bind:"+jti+":code"] = QQCode
	kv["qq:bind:"+jti+":tries"] = "0"
	kv["qq:bind:"+jti+":qq"] = qq
	err = dao.RedisDao.PipeSetKey(kv, global.LNF_CONFIG.ChenSong.BindTimeout)
	if err != nil {
		return response.CodeDatabaseError
	}
	// 判断用户是否入群
	members, err := chensong.Client.GetGroupMemberList()
	if err != nil {
		return response.CodeChenSongError
	}
	qqInGroup := false
	for _, mem := range *members {
		if strconv.FormatInt(mem.UserID, 10) == qq {
			qqInGroup = true
			break
		}
	}
	if !qqInGroup {
		return response.CodeQQUserNotInGroup
	}
	// 发送消息
	utils.LogJson("QQCode:" + QQCode)
	//res, err := chensong.Client.SendGroupMessage("[CQ:at,qq=" + qq + "] " + nickname + "，你好像在尝试绑定，我找到了验证码： " + QQCode + " 。")
	//if err != nil || res.Status != "ok" {
	//	return response.CodeChenSongError
	//}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) QQBind(id int64, jti string, reqQQ string, reqCode string) response.Code {
	// 检测表单是否符合会话
	keyCode := "qq:bind:" + jti + ":code"
	keyTries := "qq:bind:" + jti + ":tries"
	keyQQ := "qq:bind:" + jti + ":qq"
	res, err := dao.RedisDao.PipeGetString([]string{keyCode, keyTries, keyQQ})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return response.CodeQQSessionNotExist
		}
		return response.CodeDatabaseError
	}
	if res[keyCode] == "" {
		return response.CodeQQSessionNotExist
	}
	if tries, err := strconv.ParseInt(res[keyTries], 10, 64); err != nil || tries > global.LNF_CONFIG.ChenSong.BindMaxTries {
		return response.CodeQQTooManyRequests
	}
	_, err = dao.RedisDao.INCR(keyTries)
	if err != nil {
		return response.CodeDatabaseError
	}
	if reqQQ != res[keyQQ] {
		return response.CodeQQNumberError
	}
	if reqCode != res[keyCode] {
		return response.CodeQQCodeError
	}
	err = dao.RedisDao.DelKey(keyCode)
	if err != nil {
		return response.CodeDatabaseError
	}
	// 在这个函数里检测qq是否已经注册, 减少开销
	return response.CodeSuccess
}
