package banner

import (
	"fmt"
	"runtime"

	"github.com/click33/sa-token-go/core/config"
)

// Version version number | 版本号
const Version = "0.1.1"

// Banner startup banner | 启动横幅
const Banner = `
   _____         ______      __                  ______     
  / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
  \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
 ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
/____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 
                                                             
:: Sa-Token-Go ::                                 v%s
`

const (
	boxWidth      = 57
	labelWidth    = 16
	neverExpire   = "Never Expire"
	noLimit       = "No Limit"
	configured    = "*** (configured)"
	secondsFormat = "%d seconds"
)

// Print prints startup banner | 打印启动横幅
func Print() {
	fmt.Printf(Banner, Version)
	fmt.Printf(":: Go Version ::                                 %s\n", runtime.Version())
	fmt.Printf(":: GOOS/GOARCH ::                                %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

// formatConfigLine formats a configuration line with proper padding | 格式化配置行
func formatConfigLine(label string, value any) string {
	valueWidth := boxWidth - labelWidth - 5 // 57 - 16 - 5 = 36
	valueStr := fmt.Sprintf("%v", value)
	return fmt.Sprintf("│ %-*s: %-*s  │\n", labelWidth, label, valueWidth, valueStr)
}

// formatTimeout formats timeout value (seconds or special text) | 格式化超时时间值
func formatTimeout(seconds int64) string {
	if seconds > 0 {
		// Also show human-readable format for large values
		if seconds >= 86400 {
			days := seconds / 86400
			return fmt.Sprintf("%d seconds (%d days)", seconds, days)
		}
		return fmt.Sprintf(secondsFormat, seconds)
	} else if seconds == 0 {
		return neverExpire
	}
	return noLimit
}

// formatCount formats count value (number or "No Limit") | 格式化数量值
func formatCount(count int) string {
	if count > 0 {
		return fmt.Sprintf("%d", count)
	}
	return noLimit
}

// PrintWithConfig prints startup banner with full configuration | 打印启动横幅和完整配置信息
func PrintWithConfig(cfg *config.Config) {
	Print()

	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│                   Configuration                         │")
	fmt.Println("├─────────────────────────────────────────────────────────┤")

	// Token configuration | Token 配置
	fmt.Print(formatConfigLine("Token Name", cfg.TokenName))
	fmt.Print(formatConfigLine("Token Style", cfg.TokenStyle))
	fmt.Print(formatConfigLine("Token Timeout", formatTimeout(cfg.Timeout)))
	fmt.Print(formatConfigLine("Active Timeout", formatTimeout(cfg.ActiveTimeout)))

	// Login configuration | 登录配置
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Print(formatConfigLine("Auto Renew", cfg.AutoRenew))
	fmt.Print(formatConfigLine("Concurrent", cfg.IsConcurrent))
	fmt.Print(formatConfigLine("Share Token", cfg.IsShare))
	fmt.Print(formatConfigLine("Max Login Count", formatCount(cfg.MaxLoginCount)))

	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()
}
