package main

import (
	"fmt"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

func main() {
	fmt.Println("🔄 KeyPrefix 模式演示（默认带前缀 / 空前缀）")
	fmt.Println("=" + "────────────────────────────────────────────────────────────" + "=")
	fmt.Println()

	storage := memory.NewStorage()

	// 方式1: Go 默认配置（带前缀 "satoken:"）
	fmt.Println("【方式1】默认配置 - 使用前缀 'satoken:'")
	mgr1 := core.NewBuilder().
		Storage(storage).
		TokenName("satoken").  // 使用默认的 token 名称
		KeyPrefix("satoken:"). // 显式设置前缀（默认值）
		IsPrintBanner(false).
		Build()

	token1, _ := mgr1.Login("user001", "pc")
	fmt.Printf("✅ 登录成功，Token: %s\n", token1)
	fmt.Println("   Redis Keys 示例:")
	fmt.Println("   - satoken:token:" + token1)
	fmt.Println("   - satoken:account:user001:pc")
	fmt.Println("   - satoken:session:user001")
	fmt.Println()

	// 方式2: 空前缀（键名更短，与常见「无前缀」部署一致）
	fmt.Println("【方式2】空前缀 — 与使用 KeyPrefix(\"\") 的已有服务同库时常用")
	storage2 := memory.NewStorage()
	mgr2 := core.NewBuilder().
		Storage(storage2).
		TokenName("satoken"). // 与对端约定同名即可
		KeyPrefix("").        // 不加包级前缀
		IsPrintBanner(false).
		Build()

	token2, _ := mgr2.Login("user002", "web")
	fmt.Printf("✅ 登录成功，Token: %s\n", token2)
	fmt.Println("   Redis Keys 示例：")
	fmt.Println("   - token:" + token2)
	fmt.Println("   - account:user002:web")
	fmt.Println("   - session:user002")
	fmt.Println()

	// 方式3: 自定义前缀（多应用隔离）
	fmt.Println("【方式3】自定义前缀 - 用于多应用隔离")
	storage3 := memory.NewStorage()
	mgr3 := core.NewBuilder().
		Storage(storage3).
		TokenName("satoken").
		KeyPrefix("myapp:sa:"). // 自定义前缀
		IsPrintBanner(false).
		Build()

	token3, _ := mgr3.Login("user003", "app")
	fmt.Printf("✅ 登录成功，Token: %s\n", token3)
	fmt.Println("   Redis Keys 示例:")
	fmt.Println("   - myapp:sa:token:" + token3)
	fmt.Println("   - myapp:sa:account:user003:app")
	fmt.Println("   - myapp:sa:session:user003")
	fmt.Println()

	// 关键配置说明
	fmt.Println("=" + "────────────────────────────────────────────────────────────" + "=")
	fmt.Println("📝 关键配置说明:")
	fmt.Println()
	fmt.Println("1. 异构或多实例共库：")
	fmt.Println("   cfg.SetKeyPrefix(\"\")  // 设置为空字符串")
	fmt.Println("   或")
	fmt.Println("   builder.KeyPrefix(\"\")  // Builder 方式")
	fmt.Println()
	fmt.Println("2. 多应用隔离:")
	fmt.Println("   cfg.SetKeyPrefix(\"app1:\")  // 应用1")
	fmt.Println("   cfg.SetKeyPrefix(\"app2:\")  // 应用2")
	fmt.Println()
	fmt.Println("3. 默认 Go 行为:")
	fmt.Println("   cfg.SetKeyPrefix(\"satoken:\")  // 默认值")
	fmt.Println()
	fmt.Println("=" + "────────────────────────────────────────────────────────────" + "=")
}
