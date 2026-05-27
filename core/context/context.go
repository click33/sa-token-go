package context

import (
	"strings"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
)

const (
	bearerPrefix = "Bearer "
	// AuthHeaderName 标准 HTTP Authorization 头（见 RFC 7235）
	// AuthHeaderName is the standard HTTP Authorization header
	AuthHeaderName = "Authorization"
)

// SaTokenContext Sa-Token context for current request | Sa-Token上下文，用于当前请求
type SaTokenContext struct {
	ctx     adapter.RequestContext
	manager *manager.Manager
}

// NewContext creates a new Sa-Token context | 创建新的Sa-Token上下文
func NewContext(ctx adapter.RequestContext, mgr *manager.Manager) *SaTokenContext {
	return &SaTokenContext{
		ctx:     ctx,
		manager: mgr,
	}
}

// extractBearerToken 从 Authorization 头中提取 Bearer Token（大小写不敏感）
// extractBearerToken strips case-insensitive "Bearer " prefix from Authorization header
func extractBearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return ""
	}

	if len(auth) > 7 && strings.EqualFold(auth[:7], bearerPrefix) {
		return strings.TrimSpace(auth[7:])
	}

	return auth
}

// ResolveTokenName 解析本次请求应使用的 Token 键名：显式配置了非空 TokenName 则用配置，否则回退为 Authorization
// ResolveTokenName returns cfg.TokenName when set; otherwise falls back to "Authorization"
func ResolveTokenName(cfg *config.Config) string {
	if cfg != nil && strings.TrimSpace(cfg.TokenName) != "" {
		return cfg.TokenName
	}
	return AuthHeaderName
}

// ReadTokenFromRequest 从适配请求中按 Header → Cookie → Query 读取 Token，并对结果执行 CutTokenPrefix（与 PathAuth、TokenInterceptor 同源；Query 常用于 apikey）
// ReadTokenFromRequest reads token: Header first, then Cookie, then Query (apikey); applies CutTokenPrefix on the raw value
func ReadTokenFromRequest(ctx adapter.RequestContext, mgr *manager.Manager) string {
	if ctx == nil || mgr == nil {
		return ""
	}
	cfg := mgr.GetConfig()
	name := ResolveTokenName(cfg)

	readHeader := cfg == nil || cfg.IsReadHeader
	readCookie := cfg == nil || cfg.IsReadCookie

	// 1) Header：先读 TokenName 对应头；若名不是 Authorization，再尝试 Authorization Bearer 兜底
	if readHeader {
		if v := strings.TrimSpace(ctx.GetHeader(name)); v != "" {
			if strings.EqualFold(name, AuthHeaderName) {
				if t := extractBearerToken(v); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
			return mgr.CutTokenPrefix(v)
		}
		if !strings.EqualFold(name, AuthHeaderName) {
			if auth := strings.TrimSpace(ctx.GetHeader(AuthHeaderName)); auth != "" {
				if t := extractBearerToken(auth); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
		}
	}

	// 2) Cookie：键与 Header 使用同一 ResolveTokenName
	if readCookie {
		if v := strings.TrimSpace(ctx.GetCookie(name)); v != "" {
			return mgr.CutTokenPrefix(v)
		}
	}

	// 3) Query（URL 参数传递 token，例如 ?satoken=xxx 或回退名 ?Authorization=xxx）
	if v := strings.TrimSpace(ctx.GetQuery(name)); v != "" {
		return mgr.CutTokenPrefix(v)
	}

	return ""
}

// GetTokenValue gets token value from current request | 获取当前请求的Token值
func (c *SaTokenContext) GetTokenValue() string {
	return ReadTokenFromRequest(c.ctx, c.manager)
}

// IsLogin 检查当前请求是否已登录
func (c *SaTokenContext) IsLogin() bool {
	token := c.GetTokenValue()
	return c.manager.IsLogin(token)
}

// CheckLogin 检查登录（未登录抛出错误）
func (c *SaTokenContext) CheckLogin() error {
	token := c.GetTokenValue()
	return c.manager.CheckLogin(token)
}

// GetLoginID 获取当前登录ID
func (c *SaTokenContext) GetLoginID() (string, error) {
	token := c.GetTokenValue()
	return c.manager.GetLoginID(token)
}

// HasPermission 检查是否有指定权限
func (c *SaTokenContext) HasPermission(permission string) bool {
	loginID, err := c.GetLoginID()
	if err != nil {
		return false
	}
	return c.manager.HasPermission(loginID, permission)
}

// HasRole 检查是否有指定角色
func (c *SaTokenContext) HasRole(role string) bool {
	loginID, err := c.GetLoginID()
	if err != nil {
		return false
	}
	return c.manager.HasRole(loginID, role)
}

// GetRequestContext 获取原始请求上下文
func (c *SaTokenContext) GetRequestContext() adapter.RequestContext {
	return c.ctx
}

// GetManager 获取管理器
func (c *SaTokenContext) GetManager() *manager.Manager {
	return c.manager
}
