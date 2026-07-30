package cachekey

import "strings"

const (
	LoginMsgKey  = "cache:login:msg"
	LoginDataKey = "cache:login:address:data"
	LoginSalt    = "login_salt&"
)

func GenerateLoginMsgCacheKey(str string) string {
	return LoginMsgKey + ":" + strings.ToLower(str)
}

func GenerateLoginTokenCacheKey(str string) string {
	return LoginDataKey + ":" + strings.ToLower(str)
}
