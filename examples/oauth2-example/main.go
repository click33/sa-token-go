package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/gin-gonic/gin"
)

var oauth2Server *core.OAuth2Server

func main() {
	storage := memory.NewStorage()
	oauth2Server = core.NewOAuth2Server(storage)

	registerClients()

	r := gin.Default()

	r.GET("/oauth/authorize", authorizeHandler)
	r.POST("/oauth/token", tokenHandler)
	r.GET("/oauth/userinfo", userinfoHandler)
	r.POST("/oauth/revoke", revokeHandler)

	fmt.Println("OAuth2 Server running on http://localhost:8080")
	fmt.Println("\nOAuth2 Flow:")
	fmt.Println("1. GET  /oauth/authorize?client_id=webapp&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz")
	fmt.Println("2. POST /oauth/token (grant_type=authorization_code&code=...&client_id=webapp&client_secret=secret123&redirect_uri=...)")
	fmt.Println("3. GET  /oauth/userinfo (Authorization: Bearer <token>)")
	fmt.Println("4. POST /oauth/revoke (token=<token>)")

	r.Run(":8080")
}

func registerClients() {
	client := &core.OAuth2Client{
		ClientID:     "webapp",
		ClientSecret: "secret123",
		RedirectURIs: []string{
			"http://localhost:8080/callback",
			"http://localhost:3000/callback",
		},
		GrantTypes: []core.OAuth2GrantType{
			core.GrantTypeAuthorizationCode,
			core.GrantTypeRefreshToken,
		},
		Scopes: []string{"read", "write", "profile"},
	}
	oauth2Server.RegisterClient(client)

	mobileClient := &core.OAuth2Client{
		ClientID:     "mobile-app",
		ClientSecret: "mobile-secret-456",
		RedirectURIs: []string{
			"myapp://oauth/callback",
		},
		GrantTypes: []core.OAuth2GrantType{
			core.GrantTypeAuthorizationCode,
			core.GrantTypeRefreshToken,
		},
		Scopes: []string{"read", "write"},
	}
	oauth2Server.RegisterClient(mobileClient)

	fmt.Println("✅ OAuth2 Clients registered:")
	fmt.Println("  - webapp (client_id: webapp, secret: secret123)")
	fmt.Println("  - mobile-app (client_id: mobile-app, secret: mobile-secret-456)")
}

func authorizeHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	state := c.Query("state")
	scope := c.Query("scope")

	if responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_response_type"})
		return
	}

	scopes := []string{"read", "write"}
	if scope != "" {
		scopes = []string{scope}
	}

	userID := "user123"

	authCode, err := oauth2Server.GenerateAuthorizationCode(
		clientID,
		redirectURI,
		userID,
		scopes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode.Code, state)
	c.JSON(http.StatusOK, gin.H{
		"message":      "Authorization code generated",
		"code":         authCode.Code,
		"redirect_url": redirectURL,
		"user_id":      userID,
		"scopes":       scopes,
	})
}

func tokenHandler(c *gin.Context) {
	grantType := c.PostForm("grant_type")

	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeGrant(c)
	case "refresh_token":
		handleRefreshTokenGrant(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

func handleAuthorizationCodeGrant(c *gin.Context) {
	code := c.PostForm("code")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	redirectURI := c.PostForm("redirect_uri")

	accessToken, err := oauth2Server.ExchangeCodeForToken(
		code,
		clientID,
		clientSecret,
		redirectURI,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken.Token,
		"token_type":    accessToken.TokenType,
		"expires_in":    accessToken.ExpiresIn,
		"refresh_token": accessToken.RefreshToken,
		"scope":         accessToken.Scopes,
	})
}

func handleRefreshTokenGrant(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	accessToken, err := oauth2Server.RefreshAccessToken(
		refreshToken,
		clientID,
		clientSecret,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken.Token,
		"token_type":    accessToken.TokenType,
		"expires_in":    accessToken.ExpiresIn,
		"refresh_token": accessToken.RefreshToken,
		"scope":         accessToken.Scopes,
	})
}

func userinfoHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	var token string
	fmt.Sscanf(authHeader, "Bearer %s", &token)

	accessToken, err := oauth2Server.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    accessToken.UserID,
		"client_id":  accessToken.ClientID,
		"scopes":     accessToken.Scopes,
		"expires_in": accessToken.ExpiresIn,
		"issued_at":  time.Now().Unix() - accessToken.ExpiresIn,
	})
}

func revokeHandler(c *gin.Context) {
	token := c.PostForm("token")

	if err := oauth2Server.RevokeToken(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
}
