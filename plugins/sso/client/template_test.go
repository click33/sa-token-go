package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sa-tokens/sa-token-go/plugins/sso/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTicket_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sso/checkTicket" {
			http.NotFound(w, r)
			return
		}
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(map[string]string{"loginId": "user-1"})
	}))
	defer srv.Close()

	tpl := NewTemplate(&Config{ServerURL: srv.URL, ClientID: "c1"})
	id, err := tpl.CheckTicket("t1", "")
	require.NoError(t, err)
	assert.Equal(t, "user-1", id)
}

func TestCheckTicket_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	tpl := NewTemplate(&Config{ServerURL: srv.URL, ClientID: "c1"})
	_, err := tpl.CheckTicket("t1", "")
	assert.Error(t, err)
}

func TestCheckTicket_SignRejectedByServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ticket := r.FormValue("ticket")
		client := r.FormValue("client")
		sign := r.FormValue("sign")
		want := common.HMACSign("server-secret", map[string]string{"ticket": ticket, "client": client})
		if sign != want {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"loginId": "u"})
	}))
	defer srv.Close()

	tpl := NewTemplate(&Config{ServerURL: srv.URL, ClientID: "c1", SecretKey: "client-secret"})
	_, err := tpl.CheckTicket("t1", "")
	assert.Error(t, err)
}

func TestCheckTicket_InvalidBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"loginId":""}`))
	}))
	defer srv.Close()

	tpl := NewTemplate(&Config{ServerURL: srv.URL, ClientID: "c1"})
	_, err := tpl.CheckTicket("t1", "")
	assert.Error(t, err)
}
