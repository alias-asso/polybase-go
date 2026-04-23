package routes

import (
	"net/http"
	"time"

	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/golang-jwt/jwt/v5"
)

const oidcStateCookieName = "X-OIDC-State"

type oidcStateClaims struct {
	State string `json:"state"`
	jwt.RegisteredClaims
}

func createOIDCStateToken(state string, cfg *config.Config) (string, error) {
	claims := oidcStateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Auth.JWTSecret))
}

func setOIDCStateCookie(w http.ResponseWriter, state string, cfg *config.Config, isDev bool) error {
	token, err := createOIDCStateToken(state, cfg)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDev,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	return nil
}

func validOIDCState(r *http.Request, state string, cfg *config.Config) bool {
	cookie, err := r.Cookie(oidcStateCookieName)
	if err != nil {
		return false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &oidcStateClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(*oidcStateClaims)
	if !ok {
		return false
	}

	return claims.State == state
}

func clearOIDCStateCookie(w http.ResponseWriter, isDev bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDev,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
