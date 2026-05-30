package api

import (
	"context"
	"net/http"
)

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates a new account and returns a token + user.
func (c *Client) Register(ctx context.Context, username, email, password string) (*AuthResponse, error) {
	var out AuthResponse
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/auth/register",
		registerRequest{Username: username, Email: email, Password: password}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Login authenticates and returns a token + user.
func (c *Client) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	var out AuthResponse
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: email, Password: password}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
