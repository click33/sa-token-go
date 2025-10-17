package oauth2

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
)

// OAuth2 Authorization Code Flow Implementation
// OAuth2 授权码模式实现
//
// Flow | 流程:
// 1. RegisterClient() - Register OAuth2 client | 注册OAuth2客户端
// 2. GenerateAuthorizationCode() - User authorizes, get code | 用户授权，获取授权码
// 3. ExchangeCodeForToken() - Exchange code for access token | 用授权码换取访问令牌
// 4. ValidateAccessToken() - Validate access token | 验证访问令牌
// 5. RefreshAccessToken() - Use refresh token to get new token | 用刷新令牌获取新令牌
//
// Usage | 用法:
//   server := core.NewOAuth2Server(storage)
//   server.RegisterClient(&core.OAuth2Client{...})
//   authCode, _ := server.GenerateAuthorizationCode(...)
//   token, _ := server.ExchangeCodeForToken(...)

// GrantType OAuth2 grant type | OAuth2授权类型
type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code" // Authorization code flow | 授权码模式
	GrantTypeRefreshToken      GrantType = "refresh_token"      // Refresh token flow | 刷新令牌模式
	GrantTypeClientCredentials GrantType = "client_credentials" // Client credentials flow | 客户端凭证模式
	GrantTypePassword          GrantType = "password"           // Password flow | 密码模式
)

// Client OAuth2 client configuration | OAuth2客户端配置
type Client struct {
	ClientID     string      // Client ID | 客户端ID
	ClientSecret string      // Client secret | 客户端密钥
	RedirectURIs []string    // Allowed redirect URIs | 允许的回调URI
	GrantTypes   []GrantType // Allowed grant types | 允许的授权类型
	Scopes       []string    // Allowed scopes | 允许的权限范围
}

// AuthorizationCode authorization code information | 授权码信息
type AuthorizationCode struct {
	Code        string   // Authorization code | 授权码
	ClientID    string   // Client ID | 客户端ID
	RedirectURI string   // Redirect URI | 回调URI
	UserID      string   // User ID | 用户ID
	Scopes      []string // Requested scopes | 请求的权限范围
	CreateTime  int64    // Creation time | 创建时间
	ExpiresIn   int64    // Expiration time in seconds | 过期时间（秒）
	Used        bool     // Whether used | 是否已使用
}

// AccessToken access token information | 访问令牌信息
type AccessToken struct {
	Token        string   // Access token | 访问令牌
	TokenType    string   // Token type (Bearer) | 令牌类型（Bearer）
	ExpiresIn    int64    // Expiration time in seconds | 过期时间（秒）
	RefreshToken string   // Refresh token | 刷新令牌
	Scopes       []string // Granted scopes | 授予的权限范围
	UserID       string   // User ID | 用户ID
	ClientID     string   // Client ID | 客户端ID
}

// OAuth2Server OAuth2 authorization server | OAuth2授权服务器
type OAuth2Server struct {
	storage         adapter.Storage
	clients         map[string]*Client
	codeExpiration  time.Duration // Authorization code expiration (10min) | 授权码过期时间（10分钟）
	tokenExpiration time.Duration // Access token expiration (2h) | 访问令牌过期时间（2小时）
}

// NewOAuth2Server creates a new OAuth2 server | 创建新的OAuth2服务器
func NewOAuth2Server(storage adapter.Storage) *OAuth2Server {
	return &OAuth2Server{
		storage:         storage,
		clients:         make(map[string]*Client),
		codeExpiration:  10 * time.Minute, // Authorization code expires in 10 minutes | 授权码10分钟过期
		tokenExpiration: 2 * time.Hour,    // Access token expires in 2 hours | 访问令牌2小时过期
	}
}

// RegisterClient registers an OAuth2 client | 注册OAuth2客户端
func (s *OAuth2Server) RegisterClient(client *Client) {
	s.clients[client.ClientID] = client
}

// GetClient gets client by ID | 根据ID获取客户端
func (s *OAuth2Server) GetClient(clientID string) (*Client, error) {
	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}
	return client, nil
}

// GenerateAuthorizationCode generates authorization code | 生成授权码
func (s *OAuth2Server) GenerateAuthorizationCode(clientID, redirectURI, userID string, scopes []string) (*AuthorizationCode, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return nil, fmt.Errorf("invalid redirect_uri")
	}

	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(codeBytes)

	authCode := &AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		UserID:      userID,
		Scopes:      scopes,
		CreateTime:  time.Now().Unix(),
		ExpiresIn:   int64(s.codeExpiration.Seconds()),
		Used:        false,
	}

	key := fmt.Sprintf("satoken:oauth2:code:%s", code)
	if err := s.storage.Set(key, authCode, s.codeExpiration); err != nil {
		return nil, err
	}

	return authCode, nil
}

// ExchangeCodeForToken exchanges authorization code for access token | 用授权码换取访问令牌
func (s *OAuth2Server) ExchangeCodeForToken(code, clientID, clientSecret, redirectURI string) (*AccessToken, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, fmt.Errorf("invalid client credentials")
	}

	key := fmt.Sprintf("satoken:oauth2:code:%s", code)
	data, err := s.storage.Get(key)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	authCode, ok := data.(*AuthorizationCode)
	if !ok {
		return nil, fmt.Errorf("invalid code data")
	}

	if authCode.Used {
		return nil, fmt.Errorf("authorization code already used")
	}

	if authCode.ClientID != clientID {
		return nil, fmt.Errorf("client mismatch")
	}

	if authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	if time.Now().Unix() > authCode.CreateTime+authCode.ExpiresIn {
		s.storage.Delete(key)
		return nil, fmt.Errorf("authorization code expired")
	}

	authCode.Used = true
	s.storage.Set(key, authCode, time.Minute)

	return s.generateAccessToken(authCode.UserID, authCode.ClientID, authCode.Scopes)
}

func (s *OAuth2Server) generateAccessToken(userID, clientID string, scopes []string) (*AccessToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	accessToken := hex.EncodeToString(tokenBytes)

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)

	token := &AccessToken{
		Token:        accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiration.Seconds()),
		RefreshToken: refreshToken,
		Scopes:       scopes,
		UserID:       userID,
		ClientID:     clientID,
	}

	tokenKey := fmt.Sprintf("satoken:oauth2:token:%s", accessToken)
	refreshKey := fmt.Sprintf("satoken:oauth2:refresh:%s", refreshToken)

	if err := s.storage.Set(tokenKey, token, s.tokenExpiration); err != nil {
		return nil, err
	}

	if err := s.storage.Set(refreshKey, token, 30*24*time.Hour); err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateAccessToken validates access token | 验证访问令牌
func (s *OAuth2Server) ValidateAccessToken(tokenString string) (*AccessToken, error) {
	key := fmt.Sprintf("satoken:oauth2:token:%s", tokenString)
	data, err := s.storage.Get(key)
	if err != nil {
		return nil, fmt.Errorf("invalid access token")
	}

	token, ok := data.(*AccessToken)
	if !ok {
		return nil, fmt.Errorf("invalid token data")
	}

	return token, nil
}

// RefreshAccessToken refreshes access token using refresh token | 使用刷新令牌刷新访问令牌
func (s *OAuth2Server) RefreshAccessToken(refreshToken, clientID, clientSecret string) (*AccessToken, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, fmt.Errorf("invalid client credentials")
	}

	key := fmt.Sprintf("satoken:oauth2:refresh:%s", refreshToken)
	data, err := s.storage.Get(key)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	oldToken, ok := data.(*AccessToken)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token data")
	}

	if oldToken.ClientID != clientID {
		return nil, fmt.Errorf("client mismatch")
	}

	oldTokenKey := fmt.Sprintf("satoken:oauth2:token:%s", oldToken.Token)
	s.storage.Delete(oldTokenKey)

	return s.generateAccessToken(oldToken.UserID, oldToken.ClientID, oldToken.Scopes)
}

// RevokeToken revokes access token and its refresh token | 撤销访问令牌及其刷新令牌
func (s *OAuth2Server) RevokeToken(tokenString string) error {
	key := fmt.Sprintf("satoken:oauth2:token:%s", tokenString)
	data, err := s.storage.Get(key)
	if err != nil {
		return err
	}

	token, ok := data.(*AccessToken)
	if ok && token.RefreshToken != "" {
		refreshKey := fmt.Sprintf("satoken:oauth2:refresh:%s", token.RefreshToken)
		s.storage.Delete(refreshKey)
	}

	return s.storage.Delete(key)
}
