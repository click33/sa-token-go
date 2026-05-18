package gozero

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/click33/sa-token-go/core"
)

// apiErrorBody is the standard error JSON envelope for go-zero integration.
// apiErrorBody 为 go-zero 集成的标准错误 JSON 结构。
type apiErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// sentinelMatch binds a core sentinel (from errs) to business code and HTTP status.
// sentinelMatch 将 core 哨兵（源自 errs）绑定到业务码与 HTTP 状态。
type sentinelMatch struct {
	sentinel   error
	code       int
	httpStatus int
}

// sentinelMatchers order matters: more specific errors should be checked via errors.Is on wrapped chains.
// sentinelMatchers 顺序敏感；包装链上由 errors.Is 匹配。
var sentinelMatchers = []sentinelMatch{
	{core.ErrNotLogin, core.CodeNotLogin, http.StatusUnauthorized},
	{core.ErrTokenInvalid, core.CodeTokenInvalid, http.StatusUnauthorized},
	{core.ErrTokenExpired, core.CodeTokenExpired, http.StatusUnauthorized},
	{core.ErrTokenSessionTokenEmpty, core.CodeTokenInvalid, http.StatusUnauthorized},
	{core.ErrTokenSessionInvalidToken, core.CodeTokenInvalid, http.StatusUnauthorized},
	{core.ErrKickedOut, core.CodeKickedOut, http.StatusUnauthorized},
	{core.ErrTokenReplaced, core.CodeKickedOut, http.StatusUnauthorized},
	{core.ErrActiveTimeout, core.CodeActiveTimeout, http.StatusUnauthorized},
	{core.ErrSessionNotFound, core.CodeSessionError, http.StatusUnauthorized},
	{core.ErrPermissionDenied, core.CodePermissionDenied, http.StatusForbidden},
	{core.ErrRoleDenied, core.CodePermissionDenied, http.StatusForbidden},
	{core.ErrAccountDisabled, core.CodeAccountDisabled, http.StatusForbidden},
	{core.ErrDisableLevelExceeded, core.CodeAccountDisabled, http.StatusForbidden},
	{core.ErrMaxLoginCount, core.CodeMaxLoginCount, http.StatusForbidden},
	{core.ErrPathAuthRequired, core.CodePathAuthRequired, http.StatusUnauthorized},
	{core.ErrPathNotAllowed, core.CodePathNotAllowed, http.StatusForbidden},
	{core.ErrNotPassedSafeAuth, core.CodePermissionDenied, http.StatusForbidden},
	{core.ErrTokenNotFound, core.CodeNotFound, http.StatusNotFound},
	{core.ErrAccountNotFound, core.CodeNotFound, http.StatusNotFound},
	{core.ErrInvalidLoginID, core.CodeInvalidParameter, http.StatusBadRequest},
	{core.ErrInvalidDevice, core.CodeInvalidParameter, http.StatusBadRequest},
	{core.ErrInvalidConfig, core.CodeBadRequest, http.StatusBadRequest},
	{core.ErrFeatureNotSupported, core.CodeBadRequest, http.StatusBadRequest},
	{core.ErrStorageUnavailable, core.CodeStorageError, http.StatusInternalServerError},
	{core.ErrInvalidTokenData, core.CodeTokenInvalid, http.StatusUnauthorized},
	{core.ErrInvalidSessionData, core.CodeSessionError, http.StatusInternalServerError},
	{core.ErrInvalidValueType, core.CodeServerError, http.StatusInternalServerError},
}

// writeErrorResponse writes standardized JSON error using errs-based resolution.
// writeErrorResponse 按 errs 哨兵解析并写入标准化 JSON 错误响应。
func writeErrorResponse(w http.ResponseWriter, err error) {
	if err == nil {
		err = core.ErrNotLogin
	}
	code, httpStatus, message, detail := resolveAPIError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(apiErrorBody{
		Code:    code,
		Message: message,
		Error:   detail,
	})
}

// resolveAPIError maps any error to code, HTTP status, message, and detail strings.
// resolveAPIError 将任意 error 映射为 code、HTTP 状态、message、detail。
func resolveAPIError(err error) (code int, httpStatus int, message string, detail string) {
	var saErr *core.SaTokenError
	if errors.As(err, &saErr) {
		code = saErr.Code
		httpStatus = getHTTPStatusFromCode(code)
		message = saErr.Message
		if saErr.Err != nil {
			detail = saErr.Err.Error()
		} else {
			detail = saErr.Error()
		}
		return code, httpStatus, message, detail
	}

	for _, m := range sentinelMatchers {
		if errors.Is(err, m.sentinel) {
			text := m.sentinel.Error()
			return m.code, m.httpStatus, text, text
		}
	}

	code = core.CodeServerError
	httpStatus = http.StatusInternalServerError
	message = core.ErrStorageUnavailable.Error()
	detail = err.Error()
	return code, httpStatus, message, detail
}

// writeSuccessResponse writes standardized success JSON.
// writeSuccessResponse 写入标准化成功 JSON。
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    core.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// getHTTPStatusFromCode maps Sa-Token business codes to HTTP status.
// getHTTPStatusFromCode 将 Sa-Token 业务码映射为 HTTP 状态码。
func getHTTPStatusFromCode(code int) int {
	switch code {
	case core.CodeNotLogin, // same value as CodePathAuthRequired (401)
		core.CodeTokenInvalid, core.CodeTokenExpired,
		core.CodeActiveTimeout, core.CodeKickedOut, core.CodeSessionError:
		return http.StatusUnauthorized
	case core.CodePermissionDenied, // same value as CodePathNotAllowed (403)
		core.CodeAccountDisabled, core.CodeMaxLoginCount:
		return http.StatusForbidden
	case core.CodeBadRequest, core.CodeInvalidParameter:
		return http.StatusBadRequest
	case core.CodeNotFound:
		return http.StatusNotFound
	case core.CodeServerError, core.CodeStorageError:
		return http.StatusInternalServerError
	default:
		if code >= 10001 && code < 20000 {
			return http.StatusForbidden
		}
		return http.StatusInternalServerError
	}
}
