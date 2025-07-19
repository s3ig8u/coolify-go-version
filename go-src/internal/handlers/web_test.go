package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"html/template"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(string(s[0])) + s[1:]
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("../../internal/templates/**/*.html")
	return r
}

func TestRegisterPost_Success(t *testing.T) {
	r := setupTestRouter()
	h := &WebHandler{}
	r.POST("/auth/register", h.RegisterPost)

	form := url.Values{}
	form.Add("name", "Test User")
	form.Add("email", "test@example.com")
	form.Add("password", "secret123")
	form.Add("confirm-password", "secret123")
	form.Add("agree-terms", "on")

	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Register Success") {
		t.Errorf("expected success message in response")
	}
}

func TestRegisterPost_ValidationError(t *testing.T) {
	r := setupTestRouter()
	h := &WebHandler{}
	r.POST("/auth/register", h.RegisterPost)

	form := url.Values{}
	form.Add("name", "")  // missing name
	form.Add("email", "") // missing email
	form.Add("password", "abc")
	form.Add("confirm-password", "def") // mismatch
	form.Add("agree-terms", "")         // not checked

	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Full name is required") {
		t.Errorf("expected name error in response")
	}
	if !strings.Contains(w.Body.String(), "Email is required") {
		t.Errorf("expected email error in response")
	}
	if !strings.Contains(w.Body.String(), "Passwords do not match") {
		t.Errorf("expected password mismatch error in response")
	}
	if !strings.Contains(w.Body.String(), "You must agree to the terms") {
		t.Errorf("expected agree-terms error in response")
	}
}
