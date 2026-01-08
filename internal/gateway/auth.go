package gateway

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates JWT tokens
type JWTValidator struct {
	secretKey []byte
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(secretKey string) *JWTValidator {
	return &JWTValidator{
		secretKey: []byte(secretKey),
	}
}

// ValidateToken validates and extracts user ID from JWT
func (v *JWTValidator) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secretKey, nil
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return "", errors.New("token expired")
		}
		return claims.UserID, nil
	}
	
	return "", errors.New("invalid token")
}

// GenerateToken generates a JWT token (for testing)
func (v *JWTValidator) GenerateToken(userID string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(v.secretKey)
}
