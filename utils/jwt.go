package utils

/*  JWT 下线思路
 *  1, 用户特定 session 下线 jwt:blacklist:{jti}         - 1
 *  2, 用户全部下线          jwt:user:{user.ID}:version  - (int64)
 */

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

type TokenClaims struct {
	Role       int8             `mapstructure:"role"`
	FreshAfter *jwt.NumericDate `mapstructure:"buffer_time"`
	JWTVersion int64            `mapstructure:"jwt_version"`
	jwt.RegisteredClaims
}

type JWTGroup struct{}

func (j *JWTGroup) GenerateToken(user *model.User, VersionChanged bool) (string, error) {
	// 处理 version
	var (
		version int64
		err     error
	)
	if VersionChanged {
		version, err = dao.RedisDao.INCR("jwt:user:" + strconv.FormatInt(user.ID, 10) + ":version")
	} else {
		version, err = dao.RedisDao.GetValueInt64("jwt:user:" + strconv.FormatInt(user.ID, 10) + ":version")
		LogJson("jwt:user:" + strconv.FormatInt(user.ID, 10) + ":version")
		LogJson(version)
	}
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if VersionChanged {
				version = 2
			} else {
				version = 1
			}
			if err2 := dao.RedisDao.SetKey("jwt:user:"+strconv.FormatInt(user.ID, 10)+":version",
				version, global.LNF_CONFIG.JWT.ExpiresTime); err2 != nil {
				return "", err2
			}
		} else {
			return "", err
		}
	}
	// 生成 token
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Role:       user.Role,
		FreshAfter: jwt.NewNumericDate(time.Now().Add(global.LNF_CONFIG.JWT.BufferTime)),
		JWTVersion: version,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    global.LNF_CONFIG.JWT.Issuer,
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(global.LNF_CONFIG.JWT.ExpiresTime)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-30 * time.Second)),
			ID:        uuid.NewString(),
		},
	})
	return authToken.SignedString([]byte(global.LNF_CONFIG.JWT.SigningKey))
}

func (j *JWTGroup) ParseToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("InvalidSigningMethod")
		}
		return []byte(global.LNF_CONFIG.JWT.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("TokenInvalid")
}

func (j *JWTGroup) BanToken(claims *TokenClaims) error {
	if err := dao.RedisDao.SetKey("jwt:blacklist:"+claims.ID, 1, time.Until(claims.ExpiresAt.Time)); err != nil {
		return err
	}
	return nil
}

func (j *JWTGroup) IsTokenBanned(jti string) error {
	ok, err := dao.RedisDao.KeyExists("jwt:blacklist:" + jti)
	if ok && err == nil {
		return fmt.Errorf("TokenBanned")
	} else if !ok && err == nil {
		return nil
	}
	return err
}

func (j *JWTGroup) BanUserById(uid int64) error {
	_, err := dao.RedisDao.INCR("jwt:user:" + strconv.FormatInt(uid, 10) + ":version")
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("VersionNotExist")
		}
		return err
	}
	return nil
}
