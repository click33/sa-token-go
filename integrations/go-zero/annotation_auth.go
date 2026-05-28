package gozero

import (
	"net/http"
	"strings"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/stputil"
)

// hasAnyPermission returns true if loginID has any permission in list (OR semantics).
// hasAnyPermission 判断 loginID 是否拥有列表中任一权限（OR 语义）。
func hasAnyPermission(loginID string, perms []string) bool {
	for _, perm := range perms {
		if stputil.HasPermission(loginID, strings.TrimSpace(perm)) {
			return true
		}
	}
	return false
}

// hasAnyRole returns true if loginID has any role in list (OR semantics).
// hasAnyRole 判断 loginID 是否拥有列表中任一角色（OR 语义）。
func hasAnyRole(loginID string, roles []string) bool {
	for _, role := range roles {
		if stputil.HasRole(loginID, strings.TrimSpace(role)) {
			return true
		}
	}
	return false
}

// runAnnotationAuth runs annotation checks; returns updated request on success.
// runAnnotationAuth 执行注解鉴权；成功时返回更新后的 request。
func runAnnotationAuth(w http.ResponseWriter, r *http.Request, ann *Annotation) (*http.Request, bool) {
	if ann != nil && ann.Ignore {
		return r, true
	}

	rc := NewGoZeroContext(w, r)
	saCtx := core.NewContext(rc, stputil.GetManager())
	token := saCtx.GetTokenValue()

	if token == "" || !stputil.IsLogin(token) {
		writeErrorResponse(w, core.NewNotLoginError())
		return r, false
	}

	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		writeErrorResponse(w, err)
		return r, false
	}

	if ann != nil && ann.CheckDisable && stputil.IsDisable(loginID) {
		writeErrorResponse(w, core.NewAccountDisabledError(loginID))
		return r, false
	}

	if ann != nil && len(ann.CheckPermission) > 0 && !hasAnyPermission(loginID, ann.CheckPermission) {
		writeErrorResponse(w, core.NewPermissionDeniedError(strings.Join(ann.CheckPermission, ",")))
		return r, false
	}

	if ann != nil && len(ann.CheckRole) > 0 && !hasAnyRole(loginID, ann.CheckRole) {
		writeErrorResponse(w, core.NewRoleDeniedError(strings.Join(ann.CheckRole, ",")))
		return r, false
	}

	return attachSaTokenToRequest(w, r, saCtx, loginID), true
}
