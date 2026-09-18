package utils

/*  JWT 下线思路
 *  1, 用户特定 session 下线 jwt:blacklist:{jti}         - 1
 *  2, 用户全部下线          jwt:user:{user.ID}:version  - (int64)
 */

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func (j *JWTGroup) jtiKey(jti string) string {
	return "jwt:blacklist:%010d" + jti
}

func (j *JWTGroup) GenerateToken(user *model.User, VersionChanged bool) (string, error) {
	var userVersion int64
	val, err := dao.RedisDao.GetKey(fmt.Sprintf("jwt:user:%d:version", user.ID))
	if err != nil {
		return "", fmt.Errorf("ServerError")
	}
	switch v := val.(type) {
	case int64:
		userVersion = v
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "", fmt.Errorf("ServerError")
		}
		userVersion = n
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return "", fmt.Errorf("ServerError")
		}
		userVersion = n
	default:
		userVersion = 0
	}
	if VersionChanged {
		userVersion = userVersion + 1
	}
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Role:       user.Role,
		FreshAfter: jwt.NewNumericDate(time.Now().Add(global.LNF_CONFIG.JWT.BufferTime)),
		JWTVersion: userVersion,
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
		return global.LNF_CONFIG.JWT.SigningKey, nil
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
	if err := dao.RedisDao.SetKey(j.jtiKey(claims.ID), 1, time.Until(claims.ExpiresAt.Time)); err != nil {
		return err
	}
	return nil
}

func (j *JWTGroup) IsTokenBanned(jti string) error {
	ok, err := dao.RedisDao.KeyExists(j.jtiKey(jti))
	if ok && err == nil {
		return nil
	} else if !ok && err != nil {
		return fmt.Errorf("TokenBanned")
	}
	return err
}
