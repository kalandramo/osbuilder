package handler

import (
	"context"

	{{.Web.APIImportPath}}
)

// Login 用户登录.
func (h *Handler) Login(ctx context.Context, rq *{{.M.APIAlias}}.LoginRequest) (*{{.M.APIAlias}}.LoginResponse, error) {
	return h.biz.UserV1().Login(ctx, rq)
}

// RefreshToken 刷新令牌.
func (h *Handler) RefreshToken(ctx context.Context, rq *{{.M.APIAlias}}.RefreshTokenRequest) (*{{.M.APIAlias}}.RefreshTokenResponse, error) {
	return h.biz.UserV1().RefreshToken(ctx, rq)
}

// ChangePassword 修改用户密码.
func (h *Handler) ChangePassword(ctx context.Context, rq *{{.M.APIAlias}}.ChangePasswordRequest) (*{{.M.APIAlias}}.ChangePasswordResponse, error) {
	return h.biz.UserV1().ChangePassword(ctx, rq)
}

// CreateUser 创建新用户.
func (h *Handler) CreateUser(ctx context.Context, rq *{{.M.APIAlias}}.CreateUserRequest) (*{{.M.APIAlias}}.CreateUserResponse, error) {
	return h.biz.UserV1().Create(ctx, rq)
}

// UpdateUser 更新用户信息.
func (h *Handler) UpdateUser(ctx context.Context, rq *{{.M.APIAlias}}.UpdateUserRequest) (*{{.M.APIAlias}}.UpdateUserResponse, error) {
	return h.biz.UserV1().Update(ctx, rq)
}

// DeleteUser 删除用户.
func (h *Handler) DeleteUser(ctx context.Context, rq *{{.M.APIAlias}}.DeleteUserRequest) (*{{.M.APIAlias}}.DeleteUserResponse, error) {
	return h.biz.UserV1().Delete(ctx, rq)
}

// GetUser 获取用户信息.
func (h *Handler) GetUser(ctx context.Context, rq *{{.M.APIAlias}}.GetUserRequest) (*{{.M.APIAlias}}.GetUserResponse, error) {
	return h.biz.UserV1().Get(ctx, rq)
}

// ListUser 列出用户.
func (h *Handler) ListUser(ctx context.Context, rq *{{.M.APIAlias}}.ListUserRequest) (*{{.M.APIAlias}}.ListUserResponse, error) {
	return h.biz.UserV1().List(ctx, rq)
}
