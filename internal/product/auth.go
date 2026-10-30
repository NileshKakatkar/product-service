package product

import (
	"crypto/rsa"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func authenticate(path string) fiber.Handler {
	data, err := os.ReadFile(path)
	var publicKey *rsa.PublicKey
	if err == nil {
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM(data)
	}
	return func(c *fiber.Ctx) error {
		if err != nil || publicKey == nil {
			return fiber.NewError(500, "authentication configuration unavailable")
		}
		value := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(value, "Bearer ") {
			return fiber.NewError(401, "bearer token required")
		}
		claims := &Claims{}
		token, parseErr := jwt.ParseWithClaims(strings.TrimPrefix(value, "Bearer "), claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, fiber.NewError(401, "invalid signing algorithm")
			}
			return publicKey, nil
		}, jwt.WithIssuer("shopcore-user-service"), jwt.WithAudience("shopcore"))
		if parseErr != nil || token == nil || !token.Valid {
			return fiber.NewError(401, "invalid or expired token")
		}
		c.Locals("userID", claims.Subject)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func requireAdmin(c *fiber.Ctx) error {
	if c.Locals("role") != "admin" {
		return fiber.NewError(403, "administrator role required")
	}
	return c.Next()
}

func serviceAuth(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if expected == "" || c.Get("X-Service-Key") != expected {
			return fiber.NewError(401, "service authentication required")
		}
		return c.Next()
	}
}
