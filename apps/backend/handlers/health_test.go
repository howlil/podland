package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	HandleHealth(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestHandleLogin_GeneratesState(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	w := httptest.NewRecorder()

	HandleLogin(w, req)

	res := w.Result()
	defer res.Body.Close()

	// Should redirect to GitHub
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", res.StatusCode)
	}

	// Should have state cookie
	cookies := res.Cookies()
	hasStateCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "oauth_state" {
			hasStateCookie = true
			if cookie.Value == "" {
				t.Error("state cookie should not be empty")
			}
		}
	}

	if !hasStateCookie {
		t.Error("expected oauth_state cookie")
	}
}
