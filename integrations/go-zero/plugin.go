package gozero

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/click33/sa-token-go/core"
	"github.com/zeromicro/go-zero/rest"
)

// Plugin go-zero plugin for Sa-Token | go-zero插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates a go-zero plugin | 创建go-zero插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

const satokenTokenCtxKey = "satoken_token"

// TokenInterceptor parses token and injects into context for downstream handlers | 解析token并注入上下文，供后续Handler使用
func (p *Plugin) TokenInterceptor() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewGoZeroContext(w, r)
			tok := core.ReadTokenFromRequest(ctx, p.manager)
			r = r.WithContext(context.WithValue(r.Context(), satokenTokenCtxKey, tok))
			next(w, r)
		}
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewGoZeroContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				writeErrorResponse(w, err)
				return
			}

			ctx.Set("satoken", saCtx)
			next(w, r)
		}
	}
}

// PathAuthMiddleware path-based authentication middleware | 基于路径的鉴权中间件
func (p *Plugin) PathAuthMiddleware(config *core.PathAuthConfig) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			ctx := NewGoZeroContext(w, r)
			token := core.ReadTokenFromRequest(ctx, p.manager)

			result := core.ProcessAuth(path, token, config, p.manager)

			if result.ShouldReject() {
				writeErrorResponse(w, core.NewPathAuthRequiredError(path))
				return
			}

			if result.IsValid && result.TokenInfo != nil {
				ctx := NewGoZeroContext(w, r)
				saCtx := core.NewContext(ctx, p.manager)
				ctx.Set("satoken", saCtx)
				ctx.Set("loginID", result.LoginID())
			}

			next(w, r)
		}
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewGoZeroContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				writeErrorResponse(w, err)
				return
			}

			if !saCtx.HasPermission(permission) {
				writeErrorResponse(w, core.NewPermissionDeniedError(permission))
				return
			}

			ctx.Set("satoken", saCtx)
			next(w, r)
		}
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewGoZeroContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				writeErrorResponse(w, err)
				return
			}

			if !saCtx.HasRole(role) {
				writeErrorResponse(w, core.NewRoleDeniedError(role))
				return
			}

			ctx.Set("satoken", saCtx)
			next(w, r)
		}
	}
}

// LoginHandler login handler | 登录处理器
func (p *Plugin) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, core.NewError(core.CodeBadRequest, "invalid request parameters", err))
		return
	}

	device := req.Device
	if device == "" {
		device = "default"
	}

	token, err := p.manager.Login(req.Username, device)
	if err != nil {
		writeErrorResponse(w, core.NewError(core.CodeServerError, "login failed", err))
		return
	}

	writeSuccessResponse(w, map[string]interface{}{
		"token": token,
	})
}

// LogoutHandler logout handler | 登出处理器
func (p *Plugin) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewGoZeroContext(w, r)
	token := core.ReadTokenFromRequest(ctx, p.manager)

	if token == "" {
		writeErrorResponse(w, core.NewNotLoginError())
		return
	}

	if err := p.manager.LogoutByToken(token); err != nil {
		writeErrorResponse(w, err)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{
		"message": "logged out successfully",
	})
}

// UserInfoHandler user info handler | 用户信息处理器
func (p *Plugin) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewGoZeroContext(w, r)
	saCtx := core.NewContext(ctx, p.manager)

	if err := saCtx.CheckLogin(); err != nil {
		writeErrorResponse(w, err)
		return
	}

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	permissions, _ := p.manager.GetPermissions(loginID)
	roles, _ := p.manager.GetRoles(loginID)

	writeSuccessResponse(w, map[string]interface{}{
		"loginId":     loginID,
		"permissions": permissions,
		"roles":       roles,
	})
}

// GetTokenFromCtx reads token injected by TokenInterceptor from request context | 从请求context读取TokenInterceptor注入的token
func GetTokenFromCtx(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.Context().Value(satokenTokenCtxKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetSaToken retrieves Sa-Token context from request | 从请求获取Sa-Token上下文
func GetSaToken(r *http.Request) (*core.SaTokenContext, bool) {
	saToken := r.Context().Value("satoken")
	if saToken == nil {
		return nil, false
	}
	ctx, ok := saToken.(*core.SaTokenContext)
	return ctx, ok
}

// writeErrorResponse writes a standardized error response | 写入标准化错误响应
func writeErrorResponse(w http.ResponseWriter, err error) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
		"error":   err.Error(),
	})
}

// writeSuccessResponse writes a standardized success response | 写入标准化成功响应
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
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
