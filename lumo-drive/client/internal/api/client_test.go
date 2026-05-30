package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAndAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/register" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			Token: "tok123",
			User:  User{UserID: 7, Username: "neo", Email: "neo@example.com"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	resp, err := c.Register(context.Background(), "neo", "neo@example.com", "password1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.Token != "tok123" || resp.User.UserID != 7 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestErrorEnvelopeDecoded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	_, err := c.Login(context.Background(), "x@y.z", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Message != "invalid credentials" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestBearerTokenSent(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(changesResponse{Changes: []Change{{ChangeID: 5}}})
	}))
	defer srv.Close()

	c := New(srv.URL, "abc")
	changes, err := c.GetChanges(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetChanges: %v", err)
	}
	if gotAuth != "Bearer abc" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer abc")
	}
	if len(changes) != 1 || changes[0].ChangeID != 5 {
		t.Errorf("unexpected changes: %+v", changes)
	}
}

func TestMarkChunkStatusURL(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	if err := c.MarkChunkStatus(context.Background(), "file-1", 3, ChunkStatusUploaded, `"etag-xyz"`); err != nil {
		t.Fatalf("MarkChunkStatus: %v", err)
	}
	if gotPath != "/api/v1/files/file-1/chunks/3/status/uploaded" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "etag=%22etag-xyz%22" {
		t.Errorf("query = %q", gotQuery)
	}
}
