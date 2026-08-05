package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"estimulos-incentivos/internal/service"
	"estimulos-incentivos/internal/store/sqlite"
)

// testApp monta el handler completo protegido por las middlewares de seguridad,
// tal como lo arma cmd/server en producción.
type testApp struct {
	h     *Handler
	auth  *Authenticator
	store *sqlite.Store
	mux   http.Handler
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	cfg := AuthConfig{
		OperatorUser:     "op",
		OperatorPassword: "s3cret-pass",
		SessionSecret:    "0123456789abcdef0123456789abcdef",
		SessionTTL:       time.Hour,
		CookieName:       "session",
		SecureCookie:     true,
	}
	auth, err := NewAuthenticator(cfg)
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}
	dsn := filepath.Join(t.TempDir(), "test.db")
	store, err := sqlite.New(dsn)
	if err != nil {
		t.Fatalf("sqlite.New(%s): %v", dsn, err)
	}
	t.Cleanup(func() { store.Close() })

	h := New(service.New(store), auth)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return &testApp{h: h, auth: auth, store: store, mux: h.Middleware(mux)}
}

// --- request builders -------------------------------------------------------

type reqOpt func(*http.Request)

func withBody(body string) reqOpt {
	return func(r *http.Request) { r.Body = io.NopCloser(strings.NewReader(body)) }
}

func withContentType(ct string) reqOpt {
	return func(r *http.Request) { r.Header.Set("Content-Type", ct) }
}

func withCookie(c *http.Cookie) reqOpt {
	return func(r *http.Request) { r.AddCookie(c) }
}

func withHeader(k, v string) reqOpt {
	return func(r *http.Request) { r.Header.Set(k, v) }
}

func withAccept(v string) reqOpt {
	return func(r *http.Request) { r.Header.Set("Accept", v) }
}

func (a *testApp) request(t *testing.T, method, path string, opts ...reqOpt) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	for _, o := range opts {
		o(req)
	}
	rr := httptest.NewRecorder()
	a.mux.ServeHTTP(rr, req)
	return rr
}

// loginAs autentica al operador configurado y devuelve la cookie de sesión.
func loginAs(t *testing.T, a *testApp) *http.Cookie {
	t.Helper()
	rr := a.request(t, http.MethodPost, "/login",
		withBody("username=op&password=s3cret-pass"),
		withContentType("application/x-www-form-urlencoded"))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login: status %d, esperado %d", rr.Code, http.StatusSeeOther)
	}
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	t.Fatal("login: no se emitió cookie de sesión")
	return nil
}

func countEmpleados(t *testing.T, a *testApp) int {
	t.Helper()
	var n int
	if err := a.store.DB().QueryRow("SELECT count(*) FROM empleados").Scan(&n); err != nil {
		t.Fatalf("count empleados: %v", err)
	}
	return n
}

// --- Escenarios: Single-operator security boundary ---------------------------

// Spec: "Unauthenticated mutation is rejected" — sin sesión válida, la mutación
// se rechaza y NO cambia estado.
func TestUnauthenticatedMutationRejectedWithoutStateChange(t *testing.T) {
	a := newTestApp(t)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withBody(`{"nombre":"Ana","email":"ana@empresa.com","cargo":"Dev","departamento":"Ingeniería"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, esperado %d", rr.Code, http.StatusUnauthorized)
	}
	if n := countEmpleados(t, a); n != 0 {
		t.Fatalf("se persistió estado sin autenticación: %d empleados", n)
	}
}

// Una página HTML sin autenticación se redirige al login.
func TestUnauthenticatedHTMLRequestRedirectsToLogin(t *testing.T) {
	a := newTestApp(t)

	rr := a.request(t, http.MethodGet, "/", withAccept("text/html"))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status %d, esperado %d", rr.Code, http.StatusSeeOther)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Fatalf("Location %q, esperado /login", loc)
	}
}

// Spec: "Authenticated operator is allowed" — con sesión válida la ruta
// autorizada procede y la cookie usa las protecciones configuradas.
func TestLoginSuccessSetsSecureSessionCookie(t *testing.T) {
	a := newTestApp(t)

	rr := a.request(t, http.MethodPost, "/login",
		withBody("username=op&password=s3cret-pass"),
		withContentType("application/x-www-form-urlencoded"))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status %d, esperado %d", rr.Code, http.StatusSeeOther)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" {
		t.Fatalf("cookies %v, esperado solo la cookie de sesión", cookies)
	}
	c := cookies[0]
	if !c.Secure {
		t.Error("cookie sin flag Secure")
	}
	if !c.HttpOnly {
		t.Error("cookie sin flag HttpOnly")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite %v, esperado SameSiteLaxMode", c.SameSite)
	}
	if c.Path != "/" {
		t.Errorf("Path %q, esperado /", c.Path)
	}
	if c.MaxAge <= 0 {
		t.Errorf("MaxAge %d, esperado > 0", c.MaxAge)
	}
	if c.Value == "" {
		t.Error("cookie de sesión vacía")
	}
}

// Fallo de login: mensaje genérico idéntico para usuario y contraseña
// inválidos (sin enumeración de usuarios), sin cookie de sesión.
func TestLoginFailureReturnsGenericError(t *testing.T) {
	a := newTestApp(t)

	cases := []struct {
		name  string
		creds string
	}{
		{"wrong-password", "username=op&password=incorrecta"},
		{"wrong-username", "username=otro&password=s3cret-pass"},
	}
	// El mensaje debe ser idéntico en ambos casos.
	expect := `{"error":"credenciales inválidas"}`

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := a.request(t, http.MethodPost, "/login",
				withBody(tc.creds),
				withContentType("application/x-www-form-urlencoded"))
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status %d, esperado %d", rr.Code, http.StatusUnauthorized)
			}
			if body := strings.TrimSpace(rr.Body.String()); body != expect {
				t.Errorf("body %q, esperado %q", body, expect)
			}
			for _, c := range rr.Result().Cookies() {
				if c.Name == "session" {
					t.Error("login fallido emitió cookie de sesión")
				}
			}
		})
	}
}

// Operador autenticado accede a una ruta autorizada (GET de API).
func TestAuthenticatedOperatorAllowedOnAuthorizedRoute(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)

	rr := a.request(t, http.MethodGet, "/api/analisis", withCookie(cookie))

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, esperado %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type %q, esperado JSON", ct)
	}
}

// La cookie de sesión está firmada: cualquier manipulación la invalida.
func TestTamperedSessionTokenRejected(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	cookie.Value = cookie.Value + "x" // altera la firma

	rr := a.request(t, http.MethodGet, "/api/analisis", withCookie(cookie))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, esperado %d (token manipulado)", rr.Code, http.StatusUnauthorized)
	}
}

// Logout limpia la cookie de sesión en el cliente. Las sesiones son stateless
// (cookie firmada): el logout no revoca tokens ya emitidos — se documenta como
// limitación; una cookie limpiada/vacía deja de autenticar.
func TestLogoutClearsSession(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)

	rr := a.request(t, http.MethodPost, "/logout", withCookie(cookie))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("logout status %d, esperado %d", rr.Code, http.StatusSeeOther)
	}

	var cleared bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" && c.MaxAge <= 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("logout no invalidó la cookie de sesión")
	}

	// La cookie que el cliente recibe tras el logout no autentica.
	rr2 := a.request(t, http.MethodGet, "/api/analisis", withCookie(&http.Cookie{Name: "session"}))
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("cookie post-logout: status %d, esperado %d", rr2.Code, http.StatusUnauthorized)
	}
}
