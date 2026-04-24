package routes

import (
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func TestBuildOIDCAuthCodeOptionsEmpty(t *testing.T) {
	options, err := buildOIDCAuthCodeOptions("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(options) != 0 {
		t.Fatalf("expected no options, got %d", len(options))
	}
}

func TestBuildOIDCAuthCodeOptionsInvalid(t *testing.T) {
	_, err := buildOIDCAuthCodeOptions("hd=%zz")
	if err == nil {
		t.Fatal("expected parsing error, got nil")
	}
}

func TestGetOIDCURLIncludesMoreParams(t *testing.T) {
	options, err := buildOIDCAuthCodeOptions("hd=alias-asso.fr&prompt=consent")
	if err != nil {
		t.Fatalf("build options: %v", err)
	}

	s := &Server{
		oauth2Config: &oauth2.Config{
			ClientID:    "test-client",
			RedirectURL: "http://127.0.0.1:8080/auth/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://issuer.example.com/auth",
			},
		},
		oidcAuthOptions: options,
	}

	authURL, err := s.getOIDCURL("state123")
	if err != nil {
		t.Fatalf("get OIDC URL: %v", err)
	}

	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}

	query := u.Query()
	if query.Get("state") != "state123" {
		t.Fatalf("expected state=state123, got %q", query.Get("state"))
	}
	if query.Get("hd") != "alias-asso.fr" {
		t.Fatalf("expected hd=alias-asso.fr, got %q", query.Get("hd"))
	}
	if query.Get("prompt") != "consent" {
		t.Fatalf("expected prompt=consent, got %q", query.Get("prompt"))
	}
}
