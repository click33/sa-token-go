package manager

import (
	"strings"

	"github.com/sa-tokens/sa-token-go/core/errs"
)

// GetTokenName 返回配置的 Token 名称（Header/Cookie 键名）
func (m *Manager) GetTokenName() string {
	if m.config != nil && m.config.TokenName != "" {
		return m.config.TokenName
	}
	return "satoken"
}

// CutTokenPrefix 去掉 Token 前缀（宽松模式）
// 未配置前缀、或传入值不以配置前缀开头时，原样返回（不报错）
// 幂等：已裁剪过的裸 token 再次调用仍返回自身
func (m *Manager) CutTokenPrefix(raw string) string {
	if m.config == nil || m.config.TokenPrefix == "" {
		return raw
	}
	if strings.HasPrefix(raw, m.config.TokenPrefix) {
		return strings.TrimSpace(raw[len(m.config.TokenPrefix):])
	}
	return raw
}

// CutTokenPrefixStrict 去掉 Token 前缀（严格模式）
// 借鉴 Java StpLogic.getTokenValue(true)：
// - 未配置 TokenPrefix：原样返回
// - 空字符串：返回 ("", nil)，由上层按「未提交 token」处理
// - 有前缀且匹配：裁剪后返回
// - 有前缀但不匹配：返回 errs.ErrTokenNoPrefix
func (m *Manager) CutTokenPrefixStrict(raw string) (string, error) {
	if m.config == nil || m.config.TokenPrefix == "" {
		return raw, nil
	}
	if raw == "" {
		return "", nil
	}
	prefix := m.config.TokenPrefix
	if strings.HasPrefix(raw, prefix) {
		return strings.TrimSpace(raw[len(prefix):]), nil
	}
	return "", errs.ErrTokenNoPrefix
}
