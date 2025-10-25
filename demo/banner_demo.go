package main

import (
	"github.com/click33/sa-token-go/core/banner"
	"github.com/click33/sa-token-go/core/config"
)

func main() {
	// 1. 打印基础 Banner
	banner.Print()

	// 2. 打印带完整配置的 Banner
	cfg := config.DefaultConfig()
	banner.PrintWithConfig(cfg)

	// 3. 打印 JWT 配置的 Banner
	jwtCfg := &config.Config{
		TokenName:              "jwt-token",
		Timeout:                86400, // 24小时
		ActiveTimeout:          -1,
		IsConcurrent:           true,
		IsShare:                false,
		MaxLoginCount:          5,
		IsReadBody:             false,
		IsReadHeader:           true,
		IsReadCookie:           true,
		TokenStyle:             config.TokenStyleJWT,
		DataRefreshPeriod:      -1,
		TokenSessionCheckLogin: true,
		AutoRenew:              true,
		JwtSecretKey:           "my-super-secret-key-123456",
		IsLog:                  true,
		IsPrintBanner:          true,
		CookieConfig: &config.CookieConfig{
			Domain:   "example.com",
			Path:     "/api",
			Secure:   true,
			HttpOnly: true,
			SameSite: config.SameSiteStrict,
			MaxAge:   7200,
		},
	}

	banner.PrintWithConfig(jwtCfg)
}
