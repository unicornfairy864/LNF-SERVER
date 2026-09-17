package utils

/*  JWT 下线思路
 *  1, 用户特定 session 下线 jwt:blacklist:{jti}         - 1
 *  2, 用户全部下线          jwt:user:{user.ID}:version  - (int64)
 */

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

type TokenClaims struct {
	Role        int8             `mapstructure:"role"`
	BufferTime  *jwt.NumericDate `mapstructure:"buffer_time"`
	JWTVersion  int64            `mapstructure:"jwt_version"`
	jwt.RegisteredClaims
}

type JWTGroup struct{}

func (j *JWTGroup) GenerateToken(user *model.User) (string, error) {
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Role: user.Role,
		BufferTime: jwt.NewNumericDate(time.Now().Add(global.LNF_CONFIG.JWT.BufferTime)),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: global.LNF_CONFIG.JWT.Issuer,
			Subject: strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(global.LNF_CONFIG.JWT.ExpiresTime)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-30*time.Second)),
			ID: uuid.NewString(),
		},
	})
	return authToken.SignedString([]byte(global.LNF_CONFIG.JWT.SigningKey))
}

// func (j *JWTGroup) BanToken(claims TokenClaims) (string, error) {

// }