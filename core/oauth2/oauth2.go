package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	djson "github.com/click33/sa-token-go/com/codec/json"
	"github.com/click33/sa-token-go/com/storage/memory"
	"github.com/click33/sa-token-go/core/adapter"
	derror "github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/core/utils"
)

// Package oauth2 provides OAuth2 authorization server implementation OAuth2 授权服务器实现
//
// Supported Grant Types: 支持的授权类型:
//   - Authorization Code (authorization_code) 授权码模式
//   - Client Credentials (client_credentials) 客户端凭证模式
//   - Password (password) 密码模式
//   - Refresh Token (refresh_token) 刷新令牌模式
//
// Basic Flow: 基本流程:
//  1. RegisterClient() - Register OAuth2 client 注册OAuth2客户端
//  2. GenerateAuthorizationCode() - User authorizes, get code 用户授权，获取授权码
//  3. Token() or ExchangeCodeForToken() - Exchange code for access token 用授权码换取访问令牌
//  4. ValidateAccessToken() - Validate access token 验证访问令牌
//  5. RefreshAccessToken() - Use refresh token to get new token 用刷新令牌获取新令牌
//
// Usage: 用法:
//
//	server := oauth2.NewOAuth2Server(authType, prefix, storage, serializer)
//	server.RegisterClient(&oauth2.Client{...})
//	token, _ := server.Token(ctx, &oauth2.TokenRequest{...}, nil)

// Client OAuth2 client configuration OAuth2客户端配置
type Client struct {
	ClientID     string      // Client ID 客户端ID
	ClientSecret string      // Client secret 客户端密钥
	RedirectURIs []string    // Allowed redirect URIs 允许的回调URI
	GrantTypes   []GrantType // Allowed grant types 允许的授权类型
	Scopes       []string    // Allowed scopes 允许的权限范围
}

// AuthorizationCode authorization code information 授权码信息
type AuthorizationCode struct {
	Code        string   // Authorization code 授权码
	ClientID    string   // Client ID 客户端ID
	RedirectURI string   // Redirect URI 回调URI
	UserID      string   // User ID 用户ID
	Scopes      []string // Requested scopes 请求的权限范围
	CreateTime  int64    // Creation time 创建时间
	ExpiresIn   int64    // Expiration time in seconds 过期时间（秒）
	Used        bool     // Whether used 是否已使用
}

// AccessToken access token information 访问令牌信息
type AccessToken struct {
	Token        string   // Access token 访问令牌
	TokenType    string   // Token type (Bearer) 令牌类型（Bearer）
	ExpiresIn    int64    // Expiration time in seconds 过期时间（秒）
	RefreshToken string   // Refresh token 刷新令牌
	Scopes       []string // Granted scopes 授予的权限范围
	UserID       string   // User ID 用户ID
	ClientID     string   // Client ID 客户端ID
}

// TokenRequest Unified token request structure 统一的令牌请求结构
type TokenRequest struct {
	GrantType    GrantType // Required: grant type 必需：授权类型
	ClientID     string    // Required: client ID 必需：客户端ID
	ClientSecret string    // Required: client secret 必需：客户端密钥
	Code         string    // For authorization_code: authorization code 授权码模式：授权码
	RedirectURI  string    // For authorization_code: redirect URI 授权码模式：回调URI
	RefreshToken string    // For refresh_token: refresh token 刷新令牌模式：刷新令牌
	Username     string    // For password: username 密码模式：用户名
	Password     string    // For password: password 密码模式：密码
	Scopes       []string  // Optional: requested scopes 可选：请求的权限范围
}

// UserValidator Function type for validating user credentials 验证用户凭证的函数类型
type UserValidator func(username, password string) (userID string, err error)

// OAuth2Server OAuth2 authorization server OAuth2授权服务器
type OAuth2Server struct {
	authType        string             // Authentication system type 认证体系类型
	keyPrefix       string             // Configurable prefix 可配置的前缀
	clients         map[string]*Client // client map 客户端映射map
	clientsMu       sync.RWMutex       // Clients map lock 客户端映射锁
	codeExpiration  time.Duration      // Authorization code expiration (10min) 授权码过期时间（10分钟）
	tokenExpiration time.Duration      // Access token expiration (2h) 访问令牌过期时间（2小时）
	serializer      adapter.Codec      // Codec adapter for encoding and decoding operations 编解码器适配器
	storage         adapter.Storage    // Storage adapter (Redis, Memory, etc.) 存储适配器（如 Redis、Memory）
}

// NewOAuth2Server Creates a new OAuth2 server 创建新的OAuth2服务器
func NewOAuth2Server(authType, prefix string, storage adapter.Storage, serializer adapter.Codec) *OAuth2Server {
	if storage == nil {
		storage = memory.NewStorage()
	}
	if serializer == nil {
		serializer = djson.NewJSONSerializer()
	}

	return &OAuth2Server{
		authType:        authType,
		keyPrefix:       prefix,
		clients:         make(map[string]*Client),
		codeExpiration:  DefaultCodeExpiration,
		tokenExpiration: DefaultTokenExpiration,
		storage:         storage,
		serializer:      serializer,
	}
}

// RegisterClient Registers an OAuth2 client 注册OAuth2客户端
func (s *OAuth2Server) RegisterClient(client *Client) error {
	if client == nil || client.ClientID == "" {
		return derror.ErrClientOrClientIDEmpty
	}

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.clients[client.ClientID] = client
	return nil
}

// UnregisterClient Unregisters an OAuth2 client 注销OAuth2客户端
func (s *OAuth2Server) UnregisterClient(clientID string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	delete(s.clients, clientID)
}

// GetClient Gets client by ID 根据ID获取客户端
func (s *OAuth2Server) GetClient(clientID string) (*Client, error) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, derror.ErrClientNotFound
	}

	return client, nil
}

// Token Unified token endpoint that dispatches to appropriate handler based on grant type 统一的令牌端点，根据授权类型分发到相应的处理逻辑
//
// This method provides a single entry point for all OAuth2 token operations. 此方法为所有 OAuth2 令牌操作提供统一入口。
//
// Usage: 用法:
//
//	// Authorization Code Grant 授权码模式
//	token, err := server.Token(ctx, &TokenRequest{
//	    GrantType:    GrantTypeAuthorizationCode,
//	    ClientID:     "client_id",
//	    ClientSecret: "client_secret",
//	    Code:         "auth_code",
//	    RedirectURI:  "https://example.com/callback",
//	}, nil)
//
//	// Client Credentials Grant 客户端凭证模式
//	token, err := server.Token(ctx, &TokenRequest{
//	    GrantType:    GrantTypeClientCredentials,
//	    ClientID:     "client_id",
//	    ClientSecret: "client_secret",
//	    Scopes:       []string{"read", "write"},
//	}, nil)
//
//	// Password Grant 密码模式
//	token, err := server.Token(ctx, &TokenRequest{
//	    GrantType:    GrantTypePassword,
//	    ClientID:     "client_id",
//	    ClientSecret: "client_secret",
//	    Username:     "user",
//	    Password:     "pass",
//	    Scopes:       []string{"read"},
//	}, userValidator)
//
//	// Refresh Token Grant 刷新令牌模式
//	token, err := server.Token(ctx, &TokenRequest{
//	    GrantType:    GrantTypeRefreshToken,
//	    ClientID:     "client_id",
//	    ClientSecret: "client_secret",
//	    RefreshToken: "refresh_token",
//	}, nil)
func (s *OAuth2Server) Token(ctx context.Context, req *TokenRequest, validateUser UserValidator) (*AccessToken, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: token request cannot be nil", derror.ErrInvalidAuthCode)
	}

	switch req.GrantType {
	case GrantTypeAuthorizationCode:
		return s.ExchangeCodeForToken(ctx, req.Code, req.ClientID, req.ClientSecret, req.RedirectURI)

	case GrantTypeClientCredentials:
		return s.ClientCredentialsToken(ctx, req.ClientID, req.ClientSecret, req.Scopes)

	case GrantTypePassword:
		return s.PasswordGrantToken(ctx, req.ClientID, req.ClientSecret, req.Username, req.Password, req.Scopes, validateUser)

	case GrantTypeRefreshToken:
		return s.RefreshAccessToken(ctx, req.ClientID, req.RefreshToken, req.ClientSecret)

	default:
		return nil, derror.ErrInvalidGrantType
	}
}

// GenerateAuthorizationCode Generates authorization code 生成授权码
func (s *OAuth2Server) GenerateAuthorizationCode(ctx context.Context, clientID, userID, redirectURI string, scopes []string) (*AuthorizationCode, error) {
	if userID == "" {
		return nil, derror.ErrUserIDEmpty
	}

	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if !s.isValidRedirectURI(client, redirectURI) {
		return nil, derror.ErrInvalidRedirectURI
	}

	if !s.isValidScopes(client, scopes) {
		return nil, derror.ErrInvalidScope
	}

	codeBytes := make([]byte, CodeLength)
	if _, err = rand.Read(codeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random code: %w", err)
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

	encodeData, err := s.serializer.Encode(authCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	key := s.getCodeKey(code)
	if err := s.storage.Set(ctx, key, encodeData, s.codeExpiration); err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}

	return authCode, nil
}

// ExchangeCodeForToken Exchanges authorization code for access token 用授权码换取访问令牌
func (s *OAuth2Server) ExchangeCodeForToken(ctx context.Context, code, clientID, clientSecret, redirectURI string) (*AccessToken, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, derror.ErrInvalidClientCredentials
	}

	if !s.isValidGrantType(client, GrantTypeAuthorizationCode) {
		return nil, derror.ErrInvalidGrantType
	}

	key := s.getCodeKey(code)
	data, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}
	if data == nil {
		return nil, derror.ErrInvalidAuthCode
	}

	rawData, err := utils.ToBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrTypeConvert, err)
	}

	var authCode AuthorizationCode
	if err := s.serializer.Decode(rawData, &authCode); err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	if authCode.Used {
		return nil, derror.ErrAuthCodeUsed
	}

	if authCode.ClientID != clientID {
		return nil, derror.ErrClientMismatch
	}

	if authCode.RedirectURI != redirectURI {
		return nil, derror.ErrRedirectURIMismatch
	}

	if time.Now().Unix() > authCode.CreateTime+authCode.ExpiresIn {
		_ = s.storage.Delete(ctx, key)
		return nil, derror.ErrAuthCodeExpired
	}

	authCode.Used = true
	encodeData, err := s.serializer.Encode(authCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	_ = s.storage.Set(ctx, key, encodeData, time.Minute)

	return s.generateAccessToken(ctx, authCode.UserID, authCode.ClientID, authCode.Scopes)
}

// ClientCredentialsToken Gets access token using client credentials grant 使用客户端凭证模式获取访问令牌
//
// This grant type is used for server-to-server communication where no user is involved. 此授权类型用于服务器间通信，无需用户参与。
// The client authenticates with its own credentials and receives an access token. 客户端使用自己的凭证进行认证并获取访问令牌。
//
// Usage: 用法:
//
//	token, err := server.ClientCredentialsToken(ctx, "client_id", "client_secret", []string{"read", "write"})
func (s *OAuth2Server) ClientCredentialsToken(ctx context.Context, clientID, clientSecret string, scopes []string) (*AccessToken, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, derror.ErrInvalidClientCredentials
	}

	if !s.isValidGrantType(client, GrantTypeClientCredentials) {
		return nil, derror.ErrInvalidGrantType
	}

	if !s.isValidScopes(client, scopes) {
		return nil, derror.ErrInvalidScope
	}

	return s.generateAccessToken(ctx, clientID, clientID, scopes)
}

// PasswordGrantToken Gets access token using resource owner password credentials grant 使用密码模式获取访问令牌
//
// This grant type is used when the application is highly trusted (e.g., official app). 此授权类型用于高度信任的应用（如官方App）。
// The user provides their username and password directly to the client. 用户直接向客户端提供用户名和密码。
//
// SECURITY WARNING: This grant type should only be used when other flows are not viable. 安全警告：仅在其他授权流程不可行时才应使用此授权类型。
//
// Usage: 用法:
//
//	validator := func(username, password string) (string, error) {
//	    // Validate user credentials from your user store
//	    if user := userService.Authenticate(username, password); user != nil {
//	        return user.ID, nil
//	    }
//	    return "", errors.New("invalid credentials")
//	}
//	token, err := server.PasswordGrantToken(ctx, "client_id", "client_secret", "user", "pass", scopes, validator)
func (s *OAuth2Server) PasswordGrantToken(ctx context.Context, clientID, clientSecret, username, password string, scopes []string, validateUser UserValidator) (*AccessToken, error) {
	if validateUser == nil {
		return nil, fmt.Errorf("%w: user validator function is required", derror.ErrInvalidUserCredentials)
	}

	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, derror.ErrInvalidClientCredentials
	}

	if !s.isValidGrantType(client, GrantTypePassword) {
		return nil, derror.ErrInvalidGrantType
	}

	if !s.isValidScopes(client, scopes) {
		return nil, derror.ErrInvalidScope
	}

	userID, err := validateUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrInvalidUserCredentials, err)
	}

	if userID == "" {
		return nil, derror.ErrUserIDEmpty
	}

	return s.generateAccessToken(ctx, userID, clientID, scopes)
}

// RefreshAccessToken Refreshes access token using refresh token 使用刷新令牌刷新访问令牌
func (s *OAuth2Server) RefreshAccessToken(ctx context.Context, clientID, refreshToken, clientSecret string) (*AccessToken, error) {
	if refreshToken == "" {
		return nil, derror.ErrInvalidRefreshToken
	}

	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, derror.ErrInvalidClientCredentials
	}

	if !s.isValidGrantType(client, GrantTypeRefreshToken) {
		return nil, derror.ErrInvalidGrantType
	}

	key := s.getRefreshKey(refreshToken)
	data, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}
	if data == nil {
		return nil, derror.ErrInvalidRefreshToken
	}

	rawData, err := utils.ToBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrTypeConvert, err)
	}

	var accessTokenInfo AccessToken
	err = s.serializer.Decode(rawData, &accessTokenInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	if accessTokenInfo.ClientID != clientID {
		return nil, derror.ErrClientMismatch
	}

	_ = s.storage.Delete(ctx, s.getTokenKey(accessTokenInfo.Token))
	_ = s.storage.Delete(ctx, key)

	return s.generateAccessToken(ctx, accessTokenInfo.UserID, accessTokenInfo.ClientID, accessTokenInfo.Scopes)
}

// ValidateAccessToken Validates access token 验证访问令牌
func (s *OAuth2Server) ValidateAccessToken(ctx context.Context, accessToken string) bool {
	if accessToken == "" {
		return false
	}
	return s.storage.Exists(ctx, s.getTokenKey(accessToken))
}

// ValidateAccessTokenAndGetInfo Validates access token and get info 验证访问令牌并获取信息
func (s *OAuth2Server) ValidateAccessTokenAndGetInfo(ctx context.Context, accessToken string) (*AccessToken, error) {
	if accessToken == "" {
		return nil, derror.ErrInvalidAccessToken
	}

	key := s.getTokenKey(accessToken)
	data, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}
	if data == nil {
		return nil, derror.ErrInvalidAccessToken
	}

	rawData, err := utils.ToBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrTypeConvert, err)
	}

	var accessTokenInfo AccessToken
	err = s.serializer.Decode(rawData, &accessTokenInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	return &accessTokenInfo, nil
}

// RevokeToken Revokes access token and its refresh token 撤销访问令牌及其刷新令牌
func (s *OAuth2Server) RevokeToken(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil
	}

	key := s.getTokenKey(accessToken)
	data, err := s.storage.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}
	if data == nil {
		return derror.ErrInvalidAccessToken
	}

	rawData, err := utils.ToBytes(data)
	if err != nil {
		return fmt.Errorf("%w: %v", derror.ErrTypeConvert, err)
	}

	var accessTokenInfo AccessToken
	err = s.serializer.Decode(rawData, &accessTokenInfo)
	if err != nil {
		return fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	if accessTokenInfo.RefreshToken != "" {
		_ = s.storage.Delete(ctx, s.getRefreshKey(accessTokenInfo.RefreshToken))
	}

	return s.storage.Delete(ctx, key)
}

// getCodeKey Gets storage key for authorization code 获取授权码的存储键
func (s *OAuth2Server) getCodeKey(code string) string {
	return s.keyPrefix + s.authType + CodeKeySuffix + code
}

// getTokenKey Gets storage key for access token 获取访问令牌的存储键
func (s *OAuth2Server) getTokenKey(token string) string {
	return s.keyPrefix + s.authType + TokenKeySuffix + token
}

// getRefreshKey Gets storage key for refresh token 获取刷新令牌的存储键
func (s *OAuth2Server) getRefreshKey(refreshToken string) string {
	return s.keyPrefix + s.authType + RefreshKeySuffix + refreshToken
}

// isValidRedirectURI Checks if redirect URI is valid for client 检查回调URI是否有效
func (s *OAuth2Server) isValidRedirectURI(client *Client, redirectURI string) bool {
	if redirectURI == "" {
		return false
	}
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

// isValidScopes Checks if requested scopes are allowed for client 检查请求的权限范围是否被允许
func (s *OAuth2Server) isValidScopes(client *Client, scopes []string) bool {
	if len(scopes) == 0 {
		return true
	}

	if len(client.Scopes) == 0 {
		return true
	}

	allowedScopes := make(map[string]struct{}, len(client.Scopes))
	for _, scope := range client.Scopes {
		allowedScopes[scope] = struct{}{}
	}

	for _, scope := range scopes {
		if _, ok := allowedScopes[scope]; !ok {
			return false
		}
	}

	return true
}

// isValidGrantType Checks if grant type is allowed for client 检查授权类型是否被允许
func (s *OAuth2Server) isValidGrantType(client *Client, grantType GrantType) bool {
	if len(client.GrantTypes) == 0 {
		return true
	}

	for _, gt := range client.GrantTypes {
		if gt == grantType {
			return true
		}
	}
	return false
}

// generateAccessToken Generates access token and refresh token 生成访问令牌和刷新令牌
func (s *OAuth2Server) generateAccessToken(ctx context.Context, userID, clientID string, scopes []string) (*AccessToken, error) {
	tokenBytes := make([]byte, AccessTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	accessToken := hex.EncodeToString(tokenBytes)

	refreshBytes := make([]byte, RefreshTokenLength)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken := hex.EncodeToString(refreshBytes)

	token := &AccessToken{
		Token:        accessToken,
		TokenType:    TokenTypeBearer,
		ExpiresIn:    int64(s.tokenExpiration.Seconds()),
		RefreshToken: refreshToken,
		Scopes:       scopes,
		UserID:       userID,
		ClientID:     clientID,
	}

	encodeData, err := s.serializer.Encode(token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrSerializeFailed, err)
	}

	tokenKey := s.getTokenKey(accessToken)
	refreshKey := s.getRefreshKey(refreshToken)

	if err = s.storage.Set(ctx, tokenKey, encodeData, s.tokenExpiration); err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}

	if err = s.storage.Set(ctx, refreshKey, encodeData, DefaultRefreshTTL); err != nil {
		return nil, fmt.Errorf("%w: %v", derror.ErrStorageUnavailable, err)
	}

	return token, nil
}
