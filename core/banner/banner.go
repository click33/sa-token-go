package banner

import (
	"fmt"
	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"runtime"
	"strings"
	"time"
)

const (
	// BannerText stores banner text content BannerText 存储 Banner 文本内容
	BannerText = `
   _____         ______      __                  ______     
  / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
  \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
 ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
/____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 

`
)

// PrintBanner prints startup banner and key config info PrintBanner 打印启动 Banner 和关键配置信息
func PrintBanner(cfg *config.Config) {
	if cfg == nil || !cfg.IsPrintBanner {
		return
	}

	// Print banner 打印 Banner
	fmt.Print(BannerText)
	fmt.Printf(":: %-12s :: %36s\n", "Sa-Token-Go", fmt.Sprintf("v%s", core.Version))
	fmt.Printf(":: %-12s :: %36s\n", "Go Version", runtime.Version())
	fmt.Printf(":: %-12s :: %36s\n\n", "GOOS/GOARCH", runtime.GOOS+"/"+runtime.GOARCH)

	// Print config summary 打印关键配置信息
	fmt.Println("========================================")
	fmt.Println("         Configuration Summary          ")
	fmt.Println("========================================")

	// Auth config 认证配置
	fmt.Printf("AuthType         : %s\n", strings.TrimSuffix(cfg.AuthType, ":"))
	fmt.Printf("TokenName        : %s\n", cfg.TokenName)
	fmt.Printf("TokenStyle       : %s\n", getTokenStyleName(cfg.TokenStyle))

	// Timeout config 超时配置
	fmt.Printf("Timeout          : %s\n", formatDuration(cfg.Timeout))
	if cfg.AutoRenew {
		fmt.Printf("AutoRenew        : Enabled\n")
		fmt.Printf("  ├─ MaxRefresh  : %s\n", formatDuration(cfg.RenewMaxRefresh))
		fmt.Printf("  └─ Interval    : %s\n", formatDuration(cfg.RenewInterval))
	} else {
		fmt.Printf("AutoRenew        : Disabled\n")
	}
	fmt.Printf("ActiveTimeout    : %s\n", formatDuration(cfg.ActiveTimeout))

	// Concurrency config 并发配置
	fmt.Printf("Concurrency      : %s\n", formatConcurrency(cfg))

	// Token source config Token 读取配置
	fmt.Printf("Token Source     : %s\n", formatTokenSource(cfg))

	// Log config 日志配置
	if cfg.IsLog {
		fmt.Printf("Logging          : Enabled\n")
	} else {
		fmt.Printf("Logging          : Disabled\n")
	}

	fmt.Println("========================================")
	fmt.Printf("Started at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("========================================")
	fmt.Println()
}

// getTokenStyleName gets token style name getTokenStyleName 获取 Token 风格名称
func getTokenStyleName(style adapter.TokenStyle) string {
	switch style {
	case adapter.TokenStyleUUID:
		return "UUID"
	case adapter.TokenStyleSimple:
		return "Simple"
	case adapter.TokenStyleRandom32:
		return "Random-32"
	case adapter.TokenStyleRandom64:
		return "Random-64"
	case adapter.TokenStyleRandom128:
		return "Random-128"
	case adapter.TokenStyleJWT:
		return "JWT"
	case adapter.TokenStyleHash:
		return "Hash"
	case adapter.TokenStyleTimestamp:
		return "Timestamp"
	case adapter.TokenStyleTik:
		return "Tik"
	default:
		return "Unknown"
	}
}

// formatDuration formats duration display formatDuration 格式化时长显示
func formatDuration(seconds int64) string {
	if seconds == config.NoLimit {
		return "No Limit"
	}
	if seconds <= 0 {
		return "Disabled"
	}

	d := time.Duration(seconds) * time.Second

	// Day branch 天级分支
	if d >= 24*time.Hour {
		days := d / (24 * time.Hour)
		hours := (d % (24 * time.Hour)) / time.Hour
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	// Hour branch 小时级分支
	if d >= time.Hour {
		hours := d / time.Hour
		minutes := (d % time.Hour) / time.Minute
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	// Minute branch 分钟级分支
	if d >= time.Minute {
		minutes := d / time.Minute
		seconds := (d % time.Minute) / time.Second
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	// Second branch 秒级分支
	return fmt.Sprintf("%ds", seconds)
}

// formatConcurrency formats concurrency config formatConcurrency 格式化并发配置显示
func formatConcurrency(cfg *config.Config) string {
	if !cfg.IsConcurrent {
		return fmt.Sprintf("Disabled (Scope: %s)", cfg.ConcurrencyScope)
	}

	var parts []string
	parts = append(parts, "Enabled")
	parts = append(parts, fmt.Sprintf("Scope: %s", cfg.ConcurrencyScope))

	if cfg.IsShare {
		parts = append(parts, "Share: Yes")
	} else {
		parts = append(parts, "Share: No")
	}

	if cfg.MaxLoginCount == config.NoLimit {
		parts = append(parts, "Max: Unlimited")
	} else {
		parts = append(parts, fmt.Sprintf("Max: %d", cfg.MaxLoginCount))
	}

	return strings.Join(parts, ", ")
}

// formatTokenSource formats token source display formatTokenSource 格式化 Token 读取来源显示
func formatTokenSource(cfg *config.Config) string {
	var sources []string
	if cfg.IsReadHeader {
		sources = append(sources, "Header")
	}
	if cfg.IsReadCookie {
		sources = append(sources, "Cookie")
	}
	if cfg.IsReadBody {
		sources = append(sources, "Body")
	}

	if len(sources) == 0 {
		return "None"
	}

	return strings.Join(sources, ", ")
}
