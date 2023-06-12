package auth

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/golangci/golangci-lint/pkg/sliceutil"
)

type ContextKey struct {
	Key string
}

var (
	AuthUser   = ContextKey{Key: "authUser"}
	AuthHeader = ContextKey{Key: "authHeader"}
)

type UserInfo struct {
	Issuer            string         `json:"iss"`
	Active            bool           `json:"active"`
	Audience          string         `json:"aud"`
	AllowedOrigins    []string       `json:"allowed-origins,omitempty"`
	PreferredUsername string         `json:"preferred_username"`
	ID                string         `json:"id"`
	Application       string         `json:"application"`
	UserID            string         `json:"sub"`
	Email             string         `json:"email"`
	Name              string         `json:"name"`
	Username          string         `json:"username"`
	GivenName         string         `json:"given_name"`
	FamilyName        string         `json:"family_name"`
	Groups            []string       `json:"groups,omitempty"`
	RealmAccess       RealmAccess    `json:"realm_access"`
	ResourceAccess    ResourceAccess `json:"resource_access"`
	Scope             string         `json:"scope"`
	ClientID          string         `json:"client_id"`
	EmailVerified     bool           `json:"email_verified"`
	CustomerID        string         `json:"customer_id"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

type ResourceAccess struct {
	RealmManagement RealmAccess `json:"realm-management"`
	Account         RealmAccess `json:"account"`
}

func GetUserID(ctx context.Context) (string, bool) {
	info, ok := getUserInfo(ctx)
	if !ok {
		return "", ok
	}
	return info.ID, ok
}

func GetUserUUID(ctx context.Context) (uuid.UUID, error) {
	info, ok := getUserInfo(ctx)
	if !ok {
		return uuid.UUID{}, errors.New("cannot get userinfo from header")
	}
	return uuid.FromString(info.ID)
}

func GetAuthAndContentTypeHeaders(authHeader string) map[string]string {
	headerParam := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
		"x-userinfo":   authHeader,
	}
	return headerParam
}

func HasRole(ctx context.Context, roleIn string) bool {
	roles, ok := getRoles(ctx)
	if !ok {
		return false
	}
	return sliceutil.Contains(roles, roleIn)
}

func HasOneOfRoles(ctx context.Context, rolesIn []string) bool {
	roles, ok := getRoles(ctx)
	if !ok {
		return false
	}
	for _, role := range rolesIn {
		if sliceutil.Contains(roles, role) {
			return true
		}
	}
	return false
}

func NotHasRole(ctx context.Context, roleIn string) bool {
	roles, ok := getRoles(ctx)
	if !ok {
		return true
	}
	return !sliceutil.Contains(roles, roleIn)
}

func NotHasOneOfRoles(ctx context.Context, rolesIn []string) bool {
	roles, ok := getRoles(ctx)
	if !ok {
		return true
	}
	for _, role := range rolesIn {
		if sliceutil.Contains(roles, role) {
			return false
		}
	}
	return true
}

func HasRoles(ctx context.Context, rolesIn []string) bool {
	roles, ok := getRoles(ctx)
	if !ok {
		return false
	}
	for _, role := range rolesIn {
		if !sliceutil.Contains(roles, role) {
			return false
		}
	}
	return true
}

func getUserInfo(ctx context.Context) (UserInfo, bool) {
	authUser, ok := ctx.Value(AuthUser).(UserInfo)
	return authUser, ok
}

func getRoles(ctx context.Context) ([]string, bool) {
	userInfo, ok := getUserInfo(ctx)
	if !ok {
		return []string{}, false
	}
	return userInfo.RealmAccess.Roles, true
}
