package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Variables de entorno para credenciales y secreto de sesión.
// Los valores NUNCA se loguean ni se devuelven al cliente.
const (
	envOperatorUser     = "OPERATOR_USER"
	envOperatorPassword = "OPERATOR_PASSWORD"
	envSessionSecret    = "SESSION_SECRET"
	envCookieSecure     = "COOKIE_SECURE"
)

// Defaults SOLO para desarrollo. No son secretos reales: se documentan y se
// activa una advertencia en el arranque cuando se usan (sin exponer valores).
const (
	devOperatorUser     = "admin"
	devOperatorPassword = "dev-admin-password-change-me"
	devSessionSecret    = "dev-only-session-secret-change-me-0123456789abcdef"
)

// AuthConfig concentra la configuración del límite de seguridad de un operador.
type AuthConfig struct {
	OperatorUser     string
	OperatorPassword string
	SessionSecret    string
	SessionTTL       time.Duration
	CookieName       string
	SecureCookie     bool
	// DevDefaults indica que se están usando credenciales de desarrollo
	// (no aptas para producción). El arranque debe advertirlo sin exponer valores.
	DevDefaults bool
}

// AuthConfigFromEnv lee la configuración del entorno. Si faltan variables,
// aplica defaults de desarrollo documentados y marca DevDefaults.
func AuthConfigFromEnv() (AuthConfig, error) {
	cfg := AuthConfig{
		OperatorUser:     os.Getenv(envOperatorUser),
		OperatorPassword: os.Getenv(envOperatorPassword),
		SessionSecret:    os.Getenv(envSessionSecret),
		SessionTTL:       24 * time.Hour,
		CookieName:       "session",
		SecureCookie:     true,
	}
	if v := os.Getenv(envCookieSecure); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return AuthConfig{}, fmt.Errorf("%s debe ser true o false: %w", envCookieSecure, err)
		}
		cfg.SecureCookie = b
	}
	if cfg.OperatorUser == "" {
		cfg.OperatorUser = devOperatorUser
		cfg.DevDefaults = true
	}
	if cfg.OperatorPassword == "" {
		cfg.OperatorPassword = devOperatorPassword
		cfg.DevDefaults = true
	}
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = devSessionSecret
		cfg.DevDefaults = true
	}
	return cfg, nil
}

// Authenticator valida credenciales, emite/verifica sesiones firmadas y
// deriva el token CSRF ligado a cada sesión.
type Authenticator struct {
	cfg    AuthConfig
	secret []byte
}

// NewAuthenticator valida la configuración y prepara el secreto de firma.
func NewAuthenticator(cfg AuthConfig) (*Authenticator, error) {
	if cfg.OperatorUser == "" {
		return nil, errors.New("auth: OPERATOR_USER requerido")
	}
	if cfg.OperatorPassword == "" {
		return nil, errors.New("auth: OPERATOR_PASSWORD requerida")
	}
	if len(cfg.SessionSecret) < 16 {
		return nil, errors.New("auth: SESSION_SECRET debe tener al menos 16 caracteres")
	}
	if cfg.SessionTTL <= 0 {
		return nil, errors.New("auth: SessionTTL debe ser positivo")
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "session"
	}
	return &Authenticator{cfg: cfg, secret: []byte(cfg.SessionSecret)}, nil
}

// Session es la identidad autenticada dentro de una request.
type Session struct {
	Username  string
	ExpiresAt time.Time
	Token     string
}

// ValidCredentials compara usuario y contraseña en tiempo constante para no
// filtrar cuál de los dos campos era incorrecto (evita enumeración de usuarios).
func (a *Authenticator) ValidCredentials(user, pass string) bool {
	userOK := subtle.ConstantTimeCompare([]byte(user), []byte(a.cfg.OperatorUser)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(pass), []byte(a.cfg.OperatorPassword)) == 1
	return userOK && passOK
}

// NewSession emite un token de sesión firmado con expiración. El payload
// incluye un nonce aleatorio para que dos logins (incluso en el mismo segundo)
// produzcan tokens distintos.
func (a *Authenticator) NewSession() (token string, expires time.Time) {
	expires = time.Now().Add(a.cfg.SessionTTL)
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce) // crypto/rand nunca retorna error (Go >= 1.24)
	payload := strconv.FormatInt(expires.Unix(), 10) + "." + hex.EncodeToString(nonce)
	return payload + "." + a.sign(payload), expires
}

// SessionFromRequest recupera y verifica la sesión de la cookie.
func (a *Authenticator) SessionFromRequest(r *http.Request) (*Session, error) {
	c, err := r.Cookie(a.cfg.CookieName)
	if err != nil {
		return nil, err
	}
	return a.verifySession(c.Value)
}

// verifySession valida firma (tiempo constante) y expiración del token.
func (a *Authenticator) verifySession(token string) (*Session, error) {
	idx := strings.LastIndex(token, ".")
	if idx < 1 {
		return nil, errors.New("sesión malformada")
	}
	payload, sig := token[:idx], token[idx+1:]
	if subtle.ConstantTimeCompare([]byte(sig), []byte(a.sign(payload))) != 1 {
		return nil, errors.New("firma de sesión inválida")
	}
	dot := strings.Index(payload, ".")
	if dot < 1 {
		return nil, errors.New("sesión malformada")
	}
	expUnix, err := strconv.ParseInt(payload[:dot], 10, 64)
	if err != nil {
		return nil, errors.New("sesión malformada")
	}
	expires := time.Unix(expUnix, 0)
	if time.Now().After(expires) {
		return nil, errors.New("sesión expirada")
	}
	return &Session{Username: a.cfg.OperatorUser, ExpiresAt: expires, Token: token}, nil
}

// sign calcula el HMAC del dato con el secreto de sesión.
func (a *Authenticator) sign(data string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// CSRFToken deriva el token CSRF ligado a la sesión: HMAC(secreto, token de
// sesión). Al estar anclado al token de sesión, un token de otra sesión nunca
// valida (double-submit por sesión, sin cookie CSRF adicional).
func (a *Authenticator) CSRFToken(sess *Session) string {
	return a.sign(sess.Token)
}

// ValidCSRF compara el token provisto con el esperado en tiempo constante.
func (a *Authenticator) ValidCSRF(sess *Session, provided string) bool {
	if sess == nil || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(a.CSRFToken(sess))) == 1
}

// SetSessionCookie emite la cookie de sesión con Secure, HttpOnly y SameSite=Lax.
func (a *Authenticator) SetSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   a.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie invalida la cookie de sesión en el cliente.
func (a *Authenticator) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}
