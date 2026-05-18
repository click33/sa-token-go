package gozero

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/click33/sa-token-go/core"
)

func TestWriteErrorResponse_NotLoginUsesErrsText(t *testing.T) {
	w := httptest.NewRecorder()
	writeErrorResponse(w, core.NewNotLoginError())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body apiErrorBody
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, core.CodeNotLogin, body.Code)
	assert.Equal(t, "user not logged in", body.Message)
	assert.Contains(t, body.Error, core.ErrNotLogin.Error())
}

func TestWriteErrorResponse_AccountDisabledForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	writeErrorResponse(w, core.NewAccountDisabledError("u1"))
	assert.Equal(t, http.StatusForbidden, w.Code)
	var body apiErrorBody
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, core.CodeAccountDisabled, body.Code)
	assert.Contains(t, body.Error, core.ErrAccountDisabled.Error())
}

func TestWriteErrorResponse_PathAuthRequired(t *testing.T) {
	w := httptest.NewRecorder()
	writeErrorResponse(w, core.NewPathAuthRequiredError("/secure"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body apiErrorBody
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body.Error, core.ErrPathAuthRequired.Error())
}
