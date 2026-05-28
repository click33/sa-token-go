package manager

import (
	"sort"
	"strings"

	"github.com/sa-tokens/sa-token-go/core/session"
)

func sortKeywordItems(items []string, asc bool) {
	if asc {
		sort.Strings(items)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(items)))
	}
}

// SearchTokenValue scans token keys for keyword (best-effort; depends on storage Keys) | 搜索 Token
func (m *Manager) SearchTokenValue(keyword string, start, size int, asc bool) ([]string, error) {
	if size <= 0 {
		size = 10
	}
	pattern := m.prefix + TokenKeyPrefix + "*"
	keys, err := m.storage.Keys(pattern)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, k := range keys {
		tok := strings.TrimPrefix(k, m.prefix+TokenKeyPrefix)
		if keyword == "" || strings.Contains(tok, keyword) {
			out = append(out, tok)
		}
	}
	sortKeywordItems(out, asc)
	if start > 0 && start < len(out) {
		out = out[start:]
	}
	if len(out) > size {
		out = out[:size]
	}
	return out, nil
}

// SearchSessionID scans account session ids | 搜索账号 Session ID
func (m *Manager) SearchSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	if size <= 0 {
		size = 10
	}
	pattern := m.prefix + session.AccountSessionKeyPrefix + "*"
	keys, err := m.storage.Keys(pattern)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, k := range keys {
		id := strings.TrimPrefix(k, m.prefix+session.AccountSessionKeyPrefix)
		if keyword == "" || strings.Contains(id, keyword) {
			ids = append(ids, id)
		}
	}
	sortKeywordItems(ids, asc)
	if start > 0 && start < len(ids) {
		ids = ids[start:]
	}
	if len(ids) > size {
		ids = ids[:size]
	}
	return ids, nil
}

// SearchTokenSessionID scans token-session ids | 搜索 Token-Session ID
func (m *Manager) SearchTokenSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	if size <= 0 {
		size = 10
	}
	pattern := m.prefix + session.TokenSessionKeyPrefix + "*"
	keys, err := m.storage.Keys(pattern)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, k := range keys {
		id := strings.TrimPrefix(k, m.prefix+session.TokenSessionKeyPrefix)
		if keyword == "" || strings.Contains(id, keyword) {
			ids = append(ids, id)
		}
	}
	sortKeywordItems(ids, asc)
	if start > 0 && start < len(ids) {
		ids = ids[start:]
	}
	if len(ids) > size {
		ids = ids[:size]
	}
	return ids, nil
}
