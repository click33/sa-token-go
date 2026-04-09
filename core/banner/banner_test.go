package banner

import (
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"testing"
)

// TestPrintBanner_Full tests full banner config TestPrintBanner_Full 测试完整配置的 Banner 打印
func TestPrintBanner_Full(t *testing.T) {
	t.Log("========== 测试完整配置的 Banner ==========")
	cfg := &config.Config{
		IsPrintBanner: true,
		AuthType:      "login",
		TokenName:     "token",
		TokenStyle:    adapter.TokenStyleUUID,
		// Timeout sample day value Timeout 示例天级值
		Timeout:   86400,
		AutoRenew: true,
		// RenewMaxRefresh sample day value RenewMaxRefresh 示例天级值
		RenewMaxRefresh: 604800,
		// RenewInterval sample hour value RenewInterval 示例小时值
		RenewInterval: 3600,
		// ActiveTimeout sample minute value ActiveTimeout 示例分钟值
		ActiveTimeout:    1800,
		IsConcurrent:     true,
		ConcurrencyScope: "user",
		IsShare:          false,
		MaxLoginCount:    5,
		IsReadHeader:     true,
		IsReadCookie:     false,
		IsReadBody:       false,
		IsLog:            true,
	}
	PrintBanner(cfg)
}

// TestPrintBanner_Simple tests simple banner config TestPrintBanner_Simple 测试简单配置的 Banner 打印
func TestPrintBanner_Simple(t *testing.T) {
	t.Log("========== 测试简单配置的 Banner ==========")
	cfg := &config.Config{
		IsPrintBanner: true,
		AuthType:      "admin:",
		TokenName:     "admin-token",
		TokenStyle:    adapter.TokenStyleSimple,
		// Timeout sample hour value Timeout 示例小时值
		Timeout:          7200,
		AutoRenew:        false,
		ActiveTimeout:    config.NoLimit,
		IsConcurrent:     false,
		ConcurrencyScope: "device",
		IsReadHeader:     true,
		IsReadCookie:     false,
		IsReadBody:       false,
		IsLog:            false,
	}
	PrintBanner(cfg)
}

// TestPrintBanner_JWT tests JWT banner config TestPrintBanner_JWT 测试 JWT 风格的 Banner 打印
func TestPrintBanner_JWT(t *testing.T) {
	t.Log("========== 测试 JWT 风格的 Banner ==========")
	cfg := &config.Config{
		IsPrintBanner: true,
		AuthType:      "api:",
		TokenName:     "jwt-token",
		TokenStyle:    adapter.TokenStyleJWT,
		// Timeout sample hour value Timeout 示例小时值
		Timeout:   3600,
		AutoRenew: true,
		// RenewMaxRefresh sample day value RenewMaxRefresh 示例天级值
		RenewMaxRefresh: 86400,
		// RenewInterval sample minute value RenewInterval 示例分钟值
		RenewInterval: 1800,
		// ActiveTimeout sample minute value ActiveTimeout 示例分钟值
		ActiveTimeout:    600,
		IsConcurrent:     true,
		ConcurrencyScope: "user",
		IsShare:          true,
		MaxLoginCount:    config.NoLimit,
		IsReadHeader:     true,
		IsReadCookie:     true,
		IsReadBody:       true,
		IsLog:            true,
	}
	PrintBanner(cfg)
}

// TestPrintBanner_Disabled tests disabled banner TestPrintBanner_Disabled 测试禁用 Banner 打印
func TestPrintBanner_Disabled(t *testing.T) {
	t.Log("========== 测试禁用 Banner（不应该有输出）==========")
	cfg := &config.Config{
		IsPrintBanner: false,
		AuthType:      "login:",
		TokenName:     "token",
	}
	PrintBanner(cfg)
	t.Log("========== 禁用 Banner 测试完成 ==========")
}

// TestPrintBanner_Nil tests nil config TestPrintBanner_Nil 测试 nil 配置
func TestPrintBanner_Nil(t *testing.T) {
	t.Log("========== 测试 nil 配置（不应该有输出）==========")
	PrintBanner(nil)
	t.Log("========== nil 配置测试完成 ==========")
}
