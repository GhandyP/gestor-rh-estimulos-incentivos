package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// sessionCtxKey identifica la sesión autenticada dentro del contexto de request.
type sessionCtxKey struct{}

// sessionFromContext recupera la sesión autenticada, o nil si no la hay.
func sessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionCtxKey{}).(*Session)
	return s
}

// isPublicPath indica rutas accesibles sin autenticación: el formulario de
// login y el logout (logout siempre limpia la cookie aunque la sesión haya
// expirado; el CSRF de logout es un riesgo aceptado por su impacto nulo).
func isPublicPath(path string) bool {
	return path == "/login" || path == "/logout"
}

// wantsHTML distingue requests de navegador (redirección al login) de requests
// de API (respuesta 401 JSON sin exponer detalles).
func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// isMutating indica si el método puede cambiar estado (y por lo tanto exige CSRF).
func isMutating(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

// writeJSONStatus escribe una respuesta JSON con el status indicado.
func writeJSONStatus(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// RequireAuth exige una sesión de operador válida para todas las rutas salvo
// las públicas. Navegadores sin sesión se redirigen a /login; las API reciben
// un 401 genérico (mapeo de errores seguro: sin datos sensibles).
func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := a.SessionFromRequest(r)
		if err != nil {
			if wantsHTML(r) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{"error": "autenticación requerida"})
			return
		}
		ctx := context.WithValue(r.Context(), sessionCtxKey{}, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CSRFProtect exige un token CSRF válido y ligado a la sesión en toda mutación
// (métodos que cambian estado). El token se acepta por header X-CSRF-Token o
// por campo de formulario _csrf. Las rutas públicas (login/logout) y los
// métodos seguros no lo exigen. El rechazo es un 403 genérico sin detalles.
func (a *Authenticator) CSRFProtect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) || !isMutating(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		sess := sessionFromContext(r.Context())
		provided := r.Header.Get("X-CSRF-Token")
		if provided == "" {
			provided = r.FormValue("_csrf")
		}
		if !a.ValidCSRF(sess, provided) {
			writeJSONStatus(w, http.StatusForbidden, map[string]string{"error": "CSRF token inválido"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Middleware aplica la cadena de seguridad completa a las rutas registradas:
// autenticación de operador primero (deja la sesión en el contexto) y luego
// CSRF sobre las mutaciones. El orden es obligatorio: CSRFProtect necesita la
// sesión autenticada para comparar el token ligado a ella.
func (h *Handler) Middleware(next http.Handler) http.Handler {
	return h.auth.RequireAuth(h.auth.CSRFProtect(next))
}
