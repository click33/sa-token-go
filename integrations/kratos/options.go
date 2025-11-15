package kratos

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
)

// PluginOptions 认证引擎配置选项
type PluginOptions struct {
	// SkipOperations 跳过认证的operations（支持通配符）
	SkipOperations []string

	// DefaultRequireLogin 默认是否需要登录（如果没有匹配的规则，使用此默认值）
	DefaultRequireLogin bool

	// ErrorHandler 自定义错误处理器
	ErrorHandler func(ctx context.Context, err error) error
}

// defaultPluginOptions 返回默认配置
func defaultPluginOptions() *PluginOptions {
	return &PluginOptions{
		SkipOperations:      []string{},
		DefaultRequireLogin: false,
		ErrorHandler:        defaultErrorHandler,
	}
}

// defaultErrorHandler 默认错误处理器
func defaultErrorHandler(ctx context.Context, err error) error {
	// 如果已经是Kratos错误，直接返回
	if errors.IsUnauthorized(err) || errors.IsForbidden(err) {
		return err
	}

	// 根据错误类型转换为Kratos标准错误
	errMsg := err.Error()

	// 未登录错误
	if contains(errMsg, "未登录") || contains(errMsg, "token") {
		return errors.Unauthorized("UNAUTHORIZED", errMsg)
	}

	// 权限不足错误
	if contains(errMsg, "权限") || contains(errMsg, "角色") || contains(errMsg, "封禁") {
		return errors.Forbidden("FORBIDDEN", errMsg)
	}

	// 其他错误统一返回403
	return errors.Forbidden("FORBIDDEN", errMsg)
}

// ========== Option模式 ==========

// Option 配置函数
type Option func(*PluginOptions)

// WithSkipOperations 设置跳过的operations
func WithSkipOperations(operations ...string) Option {
	return func(o *PluginOptions) {
		o.SkipOperations = append(o.SkipOperations, operations...)
	}
}

// WithDefaultRequireLogin 设置默认是否需要登录
func WithDefaultRequireLogin(require bool) Option {
	return func(o *PluginOptions) {
		o.DefaultRequireLogin = require
	}
}

// WithErrorHandler 设置自定义错误处理器
func WithErrorHandler(handler func(ctx context.Context, err error) error) Option {
	return func(o *PluginOptions) {
		o.ErrorHandler = handler
	}
}
