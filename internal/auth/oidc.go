package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// OIDC handles OpenID Connect authentication
type OIDC struct {
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
	logger       *zap.Logger
}

// OIDCUserInfo represents user information from OIDC provider
type OIDCUserInfo struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Picture       string   `json:"picture"`
	Groups        []string `json:"groups"`
	Roles         []string `json:"roles"`
}

// OIDCClaims represents the claims in an OIDC token
type OIDCClaims struct {
	OIDCUserInfo
}

// NewOIDC creates a new OIDC service
func NewOIDC(issuer, clientID, clientSecret, redirectURL string, scopes []string, logger *zap.Logger) (*OIDC, error) {
	ctx := context.Background()

	// Create OIDC provider
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return &OIDC{
		provider:     provider,
		verifier:     verifier,
		oauth2Config: oauth2Config,
		logger:       logger,
	}, nil
}

// GetAuthURL generates the OIDC authorization URL
func (s *OIDC) GetAuthURL(state string) string {
	return s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for tokens
func (s *OIDC) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauth2Config.Exchange(ctx, code)
}

// ValidateIDToken validates an ID token and returns user info
func (s *OIDC) ValidateIDToken(ctx context.Context, token *oauth2.Token) (*OIDCUserInfo, error) {
	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	// Verify ID token
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims OIDCClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Convert to user info
	userInfo := &OIDCUserInfo{
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		GivenName:     claims.GivenName,
		FamilyName:    claims.FamilyName,
		Picture:       claims.Picture,
		Groups:        claims.Groups,
		Roles:         claims.Roles,
	}

	return userInfo, nil
}

// GetUserInfo fetches user information from the OIDC provider
func (s *OIDC) GetUserInfo(ctx context.Context, token *oauth2.Token) (*OIDCUserInfo, error) {
	client := s.oauth2Config.Client(ctx, token)

	// Get userinfo endpoint from provider discovery
	userInfoURL := s.provider.Endpoint().AuthURL + "/userinfo"

	// Make request to userinfo endpoint
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// RefreshToken refreshes an access token
func (s *OIDC) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	return s.oauth2Config.TokenSource(ctx, token).Token()
}

// RevokeToken revokes a token (if supported by provider)
func (s *OIDC) RevokeToken(ctx context.Context, token string) error {
	// This is provider-specific and may not be supported by all providers
	// For now, we'll just log the revocation request
	s.logger.Info("Token revocation requested", zap.String("token", token[:10]+"..."))
	return nil
}

// GenerateState generates a random state parameter for OAuth2 flow
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateState validates a state parameter
func ValidateState(expected, actual string) bool {
	return expected == actual
}

// ExtractRolesFromClaims extracts roles from OIDC claims
func ExtractRolesFromClaims(claims *OIDCClaims) []string {
	var roles []string

	// Check for roles claim
	if len(claims.Roles) > 0 {
		roles = append(roles, claims.Roles...)
	}

	// Check for groups claim (often used for roles)
	if len(claims.Groups) > 0 {
		// Map groups to roles (this is provider-specific)
		for _, group := range claims.Groups {
			if strings.HasPrefix(group, "role:") {
				roles = append(roles, strings.TrimPrefix(group, "role:"))
			} else if strings.HasPrefix(group, "mckmt-") {
				roles = append(roles, strings.TrimPrefix(group, "mckmt-"))
			}
		}
	}

	// Default role if none found
	if len(roles) == 0 {
		roles = []string{"viewer"}
	}

	return roles
}

// CreateUserFromOIDC creates an AuthenticatedUser from OIDC user info
func CreateUserFromOIDC(userInfo *OIDCUserInfo) *AuthenticatedUser {
	roles := userInfo.Roles
	if len(roles) == 0 {
		roles = ExtractRolesFromClaims(&OIDCClaims{
			OIDCUserInfo: *userInfo,
		})
	}

	return &AuthenticatedUser{
		ID:       userInfo.Subject,
		Username: userInfo.Email, // Use email as username
		Email:    userInfo.Email,
		Roles:    roles,
	}
}

// GetLogoutURL generates a logout URL (if supported by provider)
func (s *OIDC) GetLogoutURL(redirectURL string) string {
	// This is provider-specific
	// For now, we'll return a simple logout URL
	logoutURL, err := url.Parse(s.provider.Endpoint().AuthURL)
	if err != nil {
		return ""
	}

	query := logoutURL.Query()
	query.Set("logout", "true")
	if redirectURL != "" {
		query.Set("post_logout_redirect_uri", redirectURL)
	}
	logoutURL.RawQuery = query.Encode()

	return logoutURL.String()
}
