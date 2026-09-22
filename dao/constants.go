package dao

import (
	"fmt"
	"strconv"
)

var (
	qqBindPrefix = "qq:bind:"

	ConditionIDNotDeleted = "id = ? AND is_deleted = 0"
)

func GetQQBindSessionKey(qq string) string {
	return qqBindPrefix + qq
}

func GetQQBindCodeKey(jti string) string {
	return fmt.Sprintf("%s%s:code", qqBindPrefix, jti)
}

func GetQQBindTriesKey(jti string) string {
	return fmt.Sprintf("%s%s:tries", qqBindPrefix, jti)
}

func GetQQBindQQKey(jti string) string {
	return fmt.Sprintf("%s%s:qq", qqBindPrefix, jti)
}

func GetJwtVersionKey(user int64) string {
	return "jwt：user:" + strconv.FormatInt(user, 10) + ":version"
}

func GetJwtBlacklistKey(jti string) string {
	return "jwt:blacklist:" + jti
}
