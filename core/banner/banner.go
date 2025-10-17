package banner

import (
	"fmt"
	"runtime"

	"github.com/click33/sa-token-go/core/config"
)

// Version version number | 版本号
const Version = "0.1.0"

// Banner startup banner | 启动横幅
const Banner = `
   _____         ______      __                  ______     
  / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
  \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
 ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
/____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 
                                                             
:: Sa-Token-Go ::                                    (v%s)
`

// Print prints startup banner | 打印启动横幅
func Print() {
	fmt.Printf(Banner, Version)
	fmt.Printf(":: Go Version ::                                 %s\n", runtime.Version())
	fmt.Printf(":: GOOS/GOARCH ::                                %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

// PrintWithConfig prints startup banner with full configuration | 打印启动横幅和完整配置信息
func PrintWithConfig(cfg *config.Config) {
	Print()

	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│                   Configuration                         │")
	fmt.Println("├─────────────────────────────────────────────────────────┤")

	// Token configuration | Token 配置
	fmt.Printf("│ Token Name      : %-35s │\n", cfg.TokenName)
	fmt.Printf("│ Token Style     : %-35s │\n", cfg.TokenStyle)

	if cfg.Timeout > 0 {
		fmt.Printf("│ Token Timeout   : %-25d seconds │\n", cfg.Timeout)
	} else {
		fmt.Printf("│ Token Timeout   : %-35s │\n", "Never Expire")
	}

	if cfg.ActiveTimeout > 0 {
		fmt.Printf("│ Active Timeout  : %-25d seconds │\n", cfg.ActiveTimeout)
	} else {
		fmt.Printf("│ Active Timeout  : %-35s │\n", "No Limit")
	}

	// Login configuration | 登录配置
	fmt.Printf("│ Auto Renew      : %-35v │\n", cfg.AutoRenew)
	fmt.Printf("│ Concurrent      : %-35v │\n", cfg.IsConcurrent)
	fmt.Printf("│ Share Token     : %-35v │\n", cfg.IsShare)

	if cfg.MaxLoginCount > 0 {
		fmt.Printf("│ Max Login Count : %-35d │\n", cfg.MaxLoginCount)
	} else {
		fmt.Printf("│ Max Login Count : %-35s │\n", "No Limit")
	}

	// Token read source | Token 读取位置
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Read From Header: %-35v │\n", cfg.IsReadHeader)
	fmt.Printf("│ Read From Cookie: %-35v │\n", cfg.IsReadCookie)
	fmt.Printf("│ Read From Body  : %-35v │\n", cfg.IsReadBody)

	// Other settings | 其他设置
	fmt.Println("├─────────────────────────────────────────────────────────┤")

	if cfg.TokenStyle == config.TokenStyleJWT && cfg.JwtSecretKey != "" {
		fmt.Printf("│ JWT Secret      : %-35s │\n", "*** (configured)")
	}

	fmt.Printf("│ Logging         : %-35v │\n", cfg.IsLog)

	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()
}
