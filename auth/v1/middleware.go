package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func UserInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userinfoHeader := r.Header.Get("x-userinfo")
		userInfo, ok := GetUserInfoFromHeader(userinfoHeader)
		if ok {
			ctx := context.WithValue(r.Context(), AuthUser, userInfo)
			ctx = context.WithValue(ctx, AuthHeader, userinfoHeader)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserInfoFromHeader(authHeader string) (UserInfo, bool) {
	var userInfo UserInfo
	if authHeader != "" {
		decoded, err := base64.StdEncoding.DecodeString(authHeader)
		if err != nil {
			return UserInfo{}, false
		}
		err = json.Unmarshal(decoded, &userInfo)
		if err != nil {
			return UserInfo{}, false
		}
		return userInfo, true
	}
	return UserInfo{}, false
}
