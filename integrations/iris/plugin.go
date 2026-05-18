package iris

import (
	"errors"
	"net/http"

	"github.com/click33/sa-token-go/core"
	irisfw "github.com/kataras/iris/v12"
)

// Plugin Iris plugin for Sa-Token | Iris插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates an Iris plugin | 创建Iris插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// satokenTokenCtxKey TokenInterceptor 在 iris.Context 中存放解析后 token 的键
const satokenTokenCtxKey = "satoken_token"

// TokenInterceptor 解析 token 写入 iris.Context，不做登录校验
// TokenInterceptor stores parsed token on iris.Context without login check
func (p *Plugin) TokenInterceptor() irisfw.Handler {
	return func(c irisfw.Context) {
		tok := core.ReadTokenFromRequest(NewIrisContext(c), p.manager)
		c.Values().Set(satokenTokenCtxKey, tok)
		c.Next()
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() irisfw.Handler {
	return func(c irisfw.Context) {
		ctx := NewIrisContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			writeErrorResponse(c, err)
			c.StopExecution()
			return
		}

		c.Values().Set("satoken", saCtx)
		c.Next()
	}
}

// PathAuthMiddleware path-based authentication middleware | 基于路径的鉴权中间件
func (p *Plugin) PathAuthMiddleware(config *core.PathAuthConfig) irisfw.Handler {
	return func(c irisfw.Context) {
		path := c.Path()
		ctx := NewIrisContext(c)
		token := core.ReadTokenFromRequest(ctx, p.manager)

		result := core.ProcessAuth(path, token, config, p.manager)

		if result.ShouldReject() {
			writeErrorResponse(c, core.NewPathAuthRequiredError(path))
			c.StopExecution()
			return
		}

		if result.IsValid && result.TokenInfo != nil {
			ctx2 := NewIrisContext(c)
			saCtx := core.NewContext(ctx2, p.manager)
			c.Values().Set("satoken", saCtx)
			c.Values().Set("loginID", result.LoginID())
		}

		c.Next()
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) irisfw.Handler {
	return func(c irisfw.Context) {
		ctx := NewIrisContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			writeErrorResponse(c, err)
			c.StopExecution()
			return
		}

		if !saCtx.HasPermission(permission) {
			writeErrorResponse(c, core.NewPermissionDeniedError(permission))
			c.StopExecution()
			return
		}

		c.Values().Set("satoken", saCtx)
		c.Next()
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) irisfw.Handler {
	return func(c irisfw.Context) {
		ctx := NewIrisContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			writeErrorResponse(c, err)
			c.StopExecution()
			return
		}

		if !saCtx.HasRole(role) {
			writeErrorResponse(c, core.NewRoleDeniedError(role))
			c.StopExecution()
			return
		}

		c.Values().Set("satoken", saCtx)
		c.Next()
	}
}

// LoginHandler login handler example | 登录处理器示例
func (p *Plugin) LoginHandler(c irisfw.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := c.ReadJSON(&req); err != nil {
		writeErrorResponse(c, core.NewError(core.CodeBadRequest, "invalid request parameters", err))
		return
	}

	device := req.Device
	if device == "" {
		device = "default"
	}

	token, err := p.manager.Login(req.Username, device)
	if err != nil {
		writeErrorResponse(c, core.NewError(core.CodeServerError, "login failed", err))
		return
	}

	cfg := p.manager.GetConfig()
	if cfg.IsReadCookie {
		maxAge := int(cfg.Timeout)
		if maxAge < 0 {
			maxAge = 0
		}
		c.SetCookie(&http.Cookie{
			Name:     cfg.TokenName,
			Value:    token,
			MaxAge:   maxAge,
			Path:     cfg.CookieConfig.Path,
			Domain:   cfg.CookieConfig.Domain,
			Secure:   cfg.CookieConfig.Secure,
			HttpOnly: cfg.CookieConfig.HttpOnly,
		})
	}

	writeSuccessResponse(c, map[string]interface{}{
		"token": token,
	})
}

// LogoutHandler logout handler | 登出处理器
func (p *Plugin) LogoutHandler(c irisfw.Context) {
	ctx := NewIrisContext(c)
	saCtx := core.NewContext(ctx, p.manager)

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		writeErrorResponse(c, err)
		return
	}

	if err := p.manager.Logout(loginID); err != nil {
		writeErrorResponse(c, core.NewError(core.CodeServerError, "logout failed", err))
		return
	}

	writeSuccessResponse(c, map[string]interface{}{
		"message": "logout successful",
	})
}

// UserInfoHandler user info handler example | 获取用户信息处理器示例
func (p *Plugin) UserInfoHandler(c irisfw.Context) {
	ctx := NewIrisContext(c)
	saCtx := core.NewContext(ctx, p.manager)

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		writeErrorResponse(c, err)
		return
	}

	permissions, _ := p.manager.GetPermissions(loginID)
	roles, _ := p.manager.GetRoles(loginID)

	writeSuccessResponse(c, map[string]interface{}{
		"loginId":     loginID,
		"permissions": permissions,
		"roles":       roles,
	})
}

// GetTokenFromCtx 读取 TokenInterceptor 写入的 token（未挂载拦截器时返回空串）
// GetTokenFromCtx returns the token stashed by TokenInterceptor
func GetTokenFromCtx(c irisfw.Context) string {
	if v := c.Values().Get(satokenTokenCtxKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetSaToken gets Sa-Token context from Iris context | 从Iris上下文获取Sa-Token上下文
func GetSaToken(c irisfw.Context) (*core.SaTokenContext, bool) {
	satoken := c.Values().Get("satoken")
	if satoken == nil {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}

// ============ Error Handling Helpers | 错误处理辅助函数 ============

// writeErrorResponse writes a standardized error response | 写入标准化的错误响应
func writeErrorResponse(c irisfw.Context, err error) {
	var saErr *core.SaTokenError
	var code int
	var message string
	var httpStatus int

	if errors.As(err, &saErr) {
		code = saErr.Code
		message = saErr.Message
		httpStatus = getHTTPStatusFromCode(code)
	} else {
		code = core.CodeServerError
		message = err.Error()
		httpStatus = http.StatusInternalServerError
	}

	c.StatusCode(httpStatus)
	_ = c.JSON(map[string]interface{}{
		"code":    code,
		"message": message,
		"error":   err.Error(),
	})
}

// writeSuccessResponse writes a standardized success response | 写入标准化的成功响应
func writeSuccessResponse(c irisfw.Context, data interface{}) {
	c.StatusCode(http.StatusOK)
	_ = c.JSON(map[string]interface{}{
		"code":    core.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// getHTTPStatusFromCode converts Sa-Token error code to HTTP status | 将Sa-Token错误码转换为HTTP状态码
func getHTTPStatusFromCode(code int) int {
	switch code {
	case core.CodeNotLogin:
		return http.StatusUnauthorized
	case core.CodePermissionDenied:
		return http.StatusForbidden
	case core.CodeBadRequest:
		return http.StatusBadRequest
	case core.CodeNotFound:
		return http.StatusNotFound
	case core.CodeServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
