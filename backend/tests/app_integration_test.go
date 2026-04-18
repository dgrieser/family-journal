package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appbuilder "familyjournal/backend/internal/app"
	"familyjournal/backend/internal/config"
	"familyjournal/backend/internal/email"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func TestAppSetsCSRFCookieOnSafeRequests(t *testing.T) {
	app := newIntegratedTestApp(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/profile", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Set-Cookie"), "csrf_=") {
		t.Fatalf("expected csrf cookie to be issued, got %q", resp.Header.Get("Set-Cookie"))
	}
}

func TestAppRejectsUnsafeRequestsWithoutCSRFToken(t *testing.T) {
	app := newIntegratedTestApp(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"test@example.com","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestAppAllowsCredentialedCORSForConfiguredOrigins(t *testing.T) {
	const origin = "http://localhost:5173"

	app := newIntegratedTestApp(t, []string{origin})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected allow-origin %q, got %q", origin, got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentialed CORS, got %q", got)
	}
	if got := resp.Header.Get("Vary"); !strings.Contains(got, "Origin") {
		t.Fatalf("expected Vary header to include Origin, got %q", got)
	}
}

func newIntegratedTestApp(t *testing.T, origins []string) *fiber.App {
	t.Helper()

	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo, email.New(email.Config{}))
	store := session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		KeyLookup:      "cookie:fj_session",
	})

	cfg := config.Config{
		SessionSecret: "test-session-secret",
		CORSOrigins:   origins,
		UploadDir:     t.TempDir(),
		MaxUploadMB:   25,
		AllowedTypes:  []string{"image/jpeg", "image/png", "application/pdf"},
	}

	return appbuilder.New(cfg, service, store)
}
