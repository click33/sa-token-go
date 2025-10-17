package main

import (
	"fmt"
	"time"

	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
)

func main() {
	fmt.Println("Sa-Token-Go Token Styles Demo")
	fmt.Println("========================================\n")

	// Demo all token styles
	// 演示所有 Token 风格
	demoTokenStyle(core.TokenStyleUUID, "UUID Style")
	demoTokenStyle(core.TokenStyleSimple, "Simple Style")
	demoTokenStyle(core.TokenStyleRandom32, "Random32 Style")
	demoTokenStyle(core.TokenStyleRandom64, "Random64 Style")
	demoTokenStyle(core.TokenStyleRandom128, "Random128 Style")
	demoTokenStyle(core.TokenStyleJWT, "JWT Style")
	demoTokenStyle(core.TokenStyleHash, "Hash Style (SHA256)")
	demoTokenStyle(core.TokenStyleTimestamp, "Timestamp Style")
	demoTokenStyle(core.TokenStyleTik, "Tik Style (Short ID)")

	fmt.Println("\n========================================")
	fmt.Println("✅ All token styles demonstrated!")
}

func demoTokenStyle(style core.TokenStyle, name string) {
	fmt.Printf("📌 %s (%s)\n", name, style)
	fmt.Println("----------------------------------------")

	// Initialize manager with specific token style
	// 使用特定 Token 风格初始化管理器
	manager := core.NewBuilder().
		Storage(memory.NewStorage()).
		TokenStyle(style).
		Timeout(3600).
		JwtSecretKey("my-secret-key-123"). // For JWT style | 用于JWT风格
		IsPrintBanner(false).
		Build()

	stputil.SetManager(manager)

	// Generate 3 tokens to show variety
	// 生成3个Token展示多样性
	for i := 1; i <= 3; i++ {
		loginID := fmt.Sprintf("user%d", 1000+i)
		token, err := stputil.Login(loginID)
		if err != nil {
			fmt.Printf("  ❌ Error generating token: %v\n", err)
			continue
		}
		fmt.Printf("  %d. Token for %s:\n     %s\n", i, loginID, token)
	}

	// Add spacing
	fmt.Println()
	time.Sleep(10 * time.Millisecond) // Small delay to show timestamp difference
}
