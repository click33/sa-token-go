package main

import (
	"fmt"
	"time"

	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
)

func main() {
	storage := memory.NewStorage()
	manager := core.NewBuilder().
		Storage(storage).
		Timeout(3600).
		IsPrintBanner(false).
		Build()

	stputil.SetManager(manager)

	demoNonce(manager)
	fmt.Println()
	demoRefreshToken(manager)
	fmt.Println()
	demoOAuth2(manager)
}

func demoNonce(manager *core.Manager) {
	fmt.Println("=== Nonce Anti-Replay Demo ===")

	nonce, err := manager.GenerateNonce()
	if err != nil {
		fmt.Printf("Error generating nonce: %v\n", err)
		return
	}
	fmt.Printf("Generated Nonce: %s\n", nonce)

	valid := manager.VerifyNonce(nonce)
	fmt.Printf("First verification: %v (should be true)\n", valid)

	valid = manager.VerifyNonce(nonce)
	fmt.Printf("Second verification: %v (should be false - replay attack prevented)\n", valid)
}

func demoRefreshToken(manager *core.Manager) {
	fmt.Println("=== Refresh Token Demo ===")

	tokenInfo, err := manager.LoginWithRefreshToken("user1000", "web")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Access Token: %s\n", tokenInfo.AccessToken[:40]+"...")
	fmt.Printf("Refresh Token: %s\n", tokenInfo.RefreshToken[:40]+"...")
	fmt.Printf("Expires at: %s\n", time.Unix(tokenInfo.ExpireTime, 0).Format(time.RFC3339))

	fmt.Println("\nRefreshing access token...")
	newTokenInfo, err := manager.RefreshAccessToken(tokenInfo.RefreshToken)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("New Access Token: %s\n", newTokenInfo.AccessToken[:40]+"...")
	fmt.Printf("Same Refresh Token: %v\n", newTokenInfo.RefreshToken == tokenInfo.RefreshToken)
}

func demoOAuth2(manager *core.Manager) {
	fmt.Println("=== OAuth2 Authorization Code Flow Demo ===")

	oauth2Server := manager.GetOAuth2Server()

	client := &core.OAuth2Client{
		ClientID:     "webapp123",
		ClientSecret: "secret456",
		RedirectURIs: []string{"http://localhost:8080/callback"},
		GrantTypes:   []core.OAuth2GrantType{core.GrantTypeAuthorizationCode, core.GrantTypeRefreshToken},
		Scopes:       []string{"read", "write"},
	}
	oauth2Server.RegisterClient(client)
	fmt.Println("Client registered")

	authCode, err := oauth2Server.GenerateAuthorizationCode(
		"webapp123",
		"http://localhost:8080/callback",
		"user1000",
		[]string{"read", "write"},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Authorization Code: %s\n", authCode.Code[:20]+"...")

	accessToken, err := oauth2Server.ExchangeCodeForToken(
		authCode.Code,
		"webapp123",
		"secret456",
		"http://localhost:8080/callback",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Access Token: %s\n", accessToken.Token[:20]+"...")
	fmt.Printf("Token Type: %s\n", accessToken.TokenType)
	fmt.Printf("Expires In: %d seconds\n", accessToken.ExpiresIn)
	fmt.Printf("Refresh Token: %s\n", accessToken.RefreshToken[:20]+"...")
	fmt.Printf("Scopes: %v\n", accessToken.Scopes)

	validated, err := oauth2Server.ValidateAccessToken(accessToken.Token)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}
	fmt.Printf("Token validated for user: %s\n", validated.UserID)

	fmt.Println("\nRefreshing OAuth2 token...")
	newToken, err := oauth2Server.RefreshAccessToken(accessToken.RefreshToken, "webapp123", "secret456")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("New Access Token: %s\n", newToken.Token[:20]+"...")
}
