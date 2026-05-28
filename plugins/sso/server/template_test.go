package server

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serverTpl(t *testing.T) *Template {
	t.Helper()
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = true
	cfg.MaxLoginCount = -1
	m := manager.NewManager(st, cfg)
	return NewTemplate(m, &Config{
		TicketTimeout: 300,
		AllowURL:      "http://127.0.0.1/*,https://ok.example/cb",
		Clients:       map[string]*ClientCfg{"cli": {ClientID: "cli"}},
	})
}

func TestCheckRedirectURL_Wildcard(t *testing.T) {
	tpl := serverTpl(t)
	assert.NoError(t, tpl.CheckRedirectURL("http://127.0.0.1/app/callback"))
}

func TestCheckRedirectURL_Exact(t *testing.T) {
	tpl := serverTpl(t)
	assert.NoError(t, tpl.CheckRedirectURL("https://ok.example/cb"))
}

func TestCheckRedirectURL_Reject(t *testing.T) {
	tpl := serverTpl(t)
	assert.Error(t, tpl.CheckRedirectURL("https://evil.example/cb"))
}

func TestCreateTicketAndCheckTicket(t *testing.T) {
	tpl := serverTpl(t)
	tok, err := tpl.CreateTicketAndSave("cli", "u1", "tv1")
	require.NoError(t, err)
	tm, err := tpl.CheckTicketAndDelete(tok, "cli")
	require.NoError(t, err)
	assert.Equal(t, "u1", tm.LoginID)

	_, err = tpl.CheckTicketAndDelete(tok, "cli")
	assert.Error(t, err)
}

func TestCheckTicketAndDelete_ClientMismatch(t *testing.T) {
	tpl := serverTpl(t)
	tk, err := tpl.CreateTicketAndSave("cli", "u2", "")
	require.NoError(t, err)
	_, err = tpl.CheckTicketAndDelete(tk, "other")
	assert.Error(t, err)
}
