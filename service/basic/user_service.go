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

func (userService *UserServiceGroup) BatchService(ids []int64) (*[]model.PublicUserResponse, response.Code) {
	users := dao.UserDao.GetUserListByIDs(ids)
	var pubRes []model.PublicUserResponse
	for _, user := range users {
		if user.ID != 0 {
			pubRes = append(pubRes, *model.UserToPublicUserResponse(&user))
		}
	}
	return &pubRes, response.CodeSuccess
}

func (userService *UserServiceGroup) GetMeService(id int64) (*model.UserResponse, response.Code) {
	user := dao.UserDao.GetUserByID(id)
	if user.ID == 0 {
		return nil, response.CodeUserNotFoundOrBanned
	}
	return model.UserToResponse(user), response.CodeSuccess
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
	if user.Status == 0 {
		return nil, nil, response.CodeUserNotFoundOrBanned
	}
	// 生成 JWTToken
	token, err := utils.JWT.GenerateToken(user, false)
	if err != nil {
		return nil, nil, response.CodeServerError
	}
	err = dao.UserDao.UpdateUserByVK(user.ID, map[string]interface{}{"last_login_at": time.Now()})
	if err != nil {
		return nil, nil, response.CodeDatabaseError
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
	err := dao.UserDao.UpdateUserByVK(id, updates)
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) QQGetCode(qq, nickname, jti string) response.Code {
	// 绑定QQ思路: 若上一个 code 未失效, 则不生成新的
	// 判断会话是否存在
	str, err := dao.RedisDao.GetValueString(dao.GetQQBindCodeKey(jti))
	if err != nil {
		return response.CodeDatabaseError
	}
	if str != "" {
		return response.CodeQQSessionAlreadyExists
	}
	ok, err := dao.RedisDao.KeyExists(dao.GetQQBindSessionKey(qq))
	if err != nil || ok == true {
		return response.CodeQQSessionAlreadyExists
	}
	// 先生成验证码，对jti进行冷却，防止多次bot查询
	QQCode := strconv.Itoa(rand.Intn(900000) + 100000)
	kv := make(map[string]interface{})
	kv[dao.GetQQBindSessionKey(qq)] = 1
	kv[dao.GetQQBindCodeKey(jti)] = QQCode
	kv[dao.GetQQBindTriesKey(jti)] = "0"
	kv[dao.GetQQBindQQKey(jti)] = qq
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
	res, err := chensong.Client.SendGroupMessage("[CQ:at,qq=" + qq + "] " + nickname + "，你好像在尝试绑定，我找到了验证码： " + QQCode + " 。")
	if err != nil || res.Status != "ok" {
		return response.CodeChenSongError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) QQBind(id int64, jti, reqQQ, reqCode string) response.Code {
	// 检测表单是否符合会话
	keyCode := dao.GetQQBindCodeKey(jti)
	keyTries := dao.GetQQBindTriesKey(jti)
	keyQQ := dao.GetQQBindQQKey(jti)
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
	err = dao.RedisDao.DelKey(dao.GetQQBindSessionKey(reqQQ))
	if err != nil {
		return response.CodeDatabaseError
	}
	// 在这个函数里检测qq是否已经注册, 减少开销
	userQQ := dao.UserDao.GetUserByQQ(reqQQ)
	if userQQ.ID != 0 {
		return response.CodeQQAlreadyRegistered
	}
	user := dao.UserDao.GetUserByID(id)
	if user.ID == 0 || user.Status == 0 {
		return response.CodeUserNotFoundOrBanned
	}
	if user.QQ != nil {
		return response.CodeQQAlreadyRegistered
	}
	err = dao.UserDao.UpdateUserByVK(id, map[string]interface{}{"qq": reqQQ})
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) ChangeUserRoleService(req *model.ChangeUserRoleRequest) response.Code {
	if !(*req.Role == 0 || *req.Role == 1 || *req.Role == 2) {
		return response.CodeFormInvalid
	}
	err := dao.UserDao.UpdateUserByVK(req.ID, map[string]interface{}{"role": req.Role})
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (userService *UserServiceGroup) ChangeUserStatusRequest(req *model.ChangeUserStatusRequest) response.Code {
	if !(*req.Status == 0 || *req.Status == 1) {
		return response.CodeFormInvalid
	}
	if *req.Status == 1 {
		_, err := dao.RedisDao.INCR(dao.GetJwtVersionKey(req.ID))
		if err != nil {
			return response.CodeDatabaseError
		}
	}
	err := dao.UserDao.UpdateUserByVK(req.ID, map[string]interface{}{"status": req.Status})
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// ChangeUserCreditRequest
// Type: 类型: 0拾金不昧奖励 1认领成功奖励 2违规扣分 3系统调整
// Delta/Credit <= -99999 时，清零Credit
func (userService *UserServiceGroup) ChangeUserCreditRequest(req *model.AddUserCreditRequest) response.Code {
	// 指针字段防 nil：未传时取默认值（credit=0 不产生变动，type=0 拾金不昧奖励，description 空串）
	credit := int64(0)
	if req.Credit != nil {
		credit = *req.Credit
	}
	logType := int64(0)
	if req.Type != nil {
		logType = *req.Type
	}
	desc := ""
	if req.Desc != nil {
		desc = *req.Desc
	}
	err := dao.UserDao.AddUserCredit(req.ID, credit, logType, desc, req.OperatorID)
	if err != nil {
		if errors.Is(err, fmt.Errorf("credit not enough")) {
			return response.CodeCreditNotEnough
		}
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}
