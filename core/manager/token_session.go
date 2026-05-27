package manager

import (
	"errors"
	"fmt"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/core/session"
	"github.com/sa-tokens/sa-token-go/core/utils"
)

// GetTokenSession loads or optionally creates a Token-Session | 获取 Token-Session
func (m *Manager) GetTokenSession(tokenValue string, isCreate bool) (*session.Session, error) {
	if tokenValue == "" {
		return nil, errs.ErrTokenSessionTokenEmpty
	}

	sess, err := session.LoadWithKeyPrefix(tokenValue, m.storage, m.prefix, session.TokenSessionKeyPrefix)
	if err == nil {
		return sess, nil
	}
	if errors.Is(err, session.ErrSessionNotFound) {
		if !isCreate {
			return nil, nil
		}
	} else {
		return nil, err
	}

	if m.config.TokenSessionCheckLogin && !m.IsLogin(tokenValue) {
		return nil, fmt.Errorf("%w: token=%s", errs.ErrTokenSessionInvalidToken, tokenValue)
	}

	lt := m.config.EffectiveLoginType()
	newSess := session.NewTokenSession(tokenValue, m.storage, m.prefix, lt)
	if err := newSess.Persist(m.getExpiration()); err != nil {
		return nil, err
	}
	return newSess, nil
}

// GetAnonTokenSession returns anonymous token session | 匿名 Token-Session
func (m *Manager) GetAnonTokenSession(currentToken string, ctx adapter.RequestContext) (*session.Session, error) {
	if currentToken != "" {
		if s, err := session.LoadWithKeyPrefix(currentToken, m.storage, m.prefix, session.TokenSessionKeyPrefix); err == nil {
			return s, nil
		}
		if _, err := m.GetLoginIDNotCheck(currentToken); err == nil {
			s, err := m.GetTokenSession(currentToken, true)
			if err != nil {
				return nil, err
			}
			return s, nil
		}
	}

	newTok := utils.RandomString(64)
	if ctx != nil {
		ctx.Set(m.config.TokenName, newTok)
	}

	lt := m.config.EffectiveLoginType()
	s := session.NewAnonTokenSession(newTok, m.storage, m.prefix, lt)
	if err := s.Persist(m.getExpiration()); err != nil {
		return nil, err
	}
	return s, nil
}

// DeleteTokenSession removes persisted Token-Session | 删除 Token-Session
func (m *Manager) DeleteTokenSession(tokenValue string) error {
	if tokenValue == "" {
		return nil
	}
	key := m.prefix + session.TokenSessionKeyPrefix + tokenValue
	return m.storage.Delete(key)
}
