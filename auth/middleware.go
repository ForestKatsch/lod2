package auth

import "context"

const UserInfoContextKey = "userInfo"

type UserInfo struct {
	UserId   string
	Username string
}

func GetCurrentUserInfo(ctx context.Context) *UserInfo {
	if ctx == nil {
		return nil
	}

	userInfo, ok := ctx.Value(UserInfoContextKey).(UserInfo)

	if !ok {
		return nil
	}

	return &userInfo
}

func IsUserLoggedIn(ctx context.Context) bool {
	if GetCurrentUserInfo(ctx) == nil {
		return false
	}

	return true
}
