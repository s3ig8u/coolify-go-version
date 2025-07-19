package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT configuration
const (
	AccessTokenExpiry  = 15 * time.Minute   // Short-lived access tokens
	RefreshTokenExpiry = 7 * 24 * time.Hour // Long-lived refresh tokens
)

// Claims represents the JWT claims
type Claims struct {
	UserID  uuid.UUID  `json:"user_id"`
	Email   string     `json:"email"`
	TeamID  *uuid.UUID `json:"team_id,omitempty"`
	Role    string     `json:"role"`
	TokenID uuid.UUID  `json:"jti"` // JWT ID for token revocation
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		accessExpiry:  AccessTokenExpiry,
		refreshExpiry: RefreshTokenExpiry,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (j *JWTManager) GenerateTokenPair(userID uuid.UUID, email string, teamID *uuid.UUID, role string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessTokenID := uuid.New()
	accessClaims := &Claims{
		UserID:  userID,
		Email:   email,
		TeamID:  teamID,
		Role:    role,
		TokenID: accessTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "coolify-go",
			Subject:   userID.String(),
			ID:        accessTokenID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshTokenID := uuid.New()
	refreshClaims := &Claims{
		UserID:  userID,
		Email:   email,
		TeamID:  teamID,
		Role:    role,
		TokenID: refreshTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "coolify-go",
			Subject:   userID.String(),
			ID:        refreshTokenID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(j.accessExpiry.Seconds()),
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshTokenPair generates a new token pair using a valid refresh token
func (j *JWTManager) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.TeamID, claims.Role)
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check for Bearer token format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[7:], nil
}

// Common JWT errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)
