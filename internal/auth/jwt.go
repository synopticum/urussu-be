// Package auth implements JWT authentication: token generation/parsing,
// the HTTP middleware guarding the API, and the context helpers carrying
// the authenticated identity to the layers below.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// claims is the JWT payload: registered claims plus the user's role.
type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Generate signs an HS256 token carrying the user's ID (sub), role, issue
// time (iat) and expiry (exp = now + ttl).
func Generate(secret, userID, role string, ttl time.Duration) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// Parse validates the token's signature and expiry and returns the user ID
// and role it carries. An expired token yields an error wrapping
// jwt.ErrTokenExpired, so callers can distinguish it with errors.Is.
func Parse(secret, tokenString string) (userID, role string, err error) {
	var c claims

	_, err = jwt.ParseWithClaims(tokenString, &c, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return "", "", err
	}

	// WithExpirationRequired covers exp, but sub and role are only validated
	// by the library when explicitly asked to — require them here.
	if c.Subject == "" || c.Role == "" {
		return "", "", errors.New("token missing sub or role claim")
	}

	return c.Subject, c.Role, nil
}

// IsExpired reports whether err means the token is past its expiry.
func IsExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}
