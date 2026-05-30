package main

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

const DEFAULT_OWNER_ID = 1

// getJWTSecret returns the signing key used for issuing and validating tokens.
// It is read from the JWT_SECRET environment variable, falling back to a
// development-only default so the server still runs locally without config.
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
	}
	return []byte(secret)
}

func AuthMiddleware(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")

	// Accept both "Bearer <token>" and a bare token.
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	if tokenString == "" {
		return fiber.ErrUnauthorized
	}

	// Validate token, enforcing the expected HMAC signing method.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return getJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		return fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.ErrUnauthorized
	}

	rawUserID, ok := claims["user_id"].(float64)
	if !ok {
		return fiber.ErrUnauthorized
	}

	c.Locals("user_id", int64(rawUserID))

	return c.Next()
}

func getUserID(c *fiber.Ctx) int64 {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return DEFAULT_OWNER_ID
	}
	return userID
}
