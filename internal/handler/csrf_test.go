package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// seedEmpleado persiste un empleado completo (empleado + perfil MAP + umbral)
// vía service, como en el flujo real, para ejercitar mutaciones que requieren
// una entidad operativa (p.ej. recomendar).
func seedEmpleado(t *testing.T, a *testApp) int64 {
	t.Helper()
	e, err := a.h.Svc.CreateEmpleado(context.Background(), "Ana", "ana@empresa.com", "Dev", "Ingeniería")
	if err != nil {
		t.Fatalf("seed empleado: %v", err)
	}
	return e.ID
}

// csrfFor deriva el token CSRF ligado a la sesión representada por la cookie.
// En el flujo real el token llega al navegador vía meta tag / hidden input;
// aquí se deriva por el mismo mecanismo de sesión que valida el middleware.
func csrfFor(t *testing.T, a *testApp, cookie *http.Cookie) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	sess, err := a.auth.SessionFromRequest(req)
	if err != nil {
		t.Fatalf("sesión desde cookie: %v", err)
	}
	return a.auth.CSRFToken(sess)
}

// --- Escenarios: CSRF-protected browser mutations ----------------------------

// Spec: "Invalid token is rejected" — token ausente: 403 y sin persistencia.
func TestCSRFAbsentRejectedWithoutStateChange(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withCookie(cookie),
		withBody(`{"nombre":"Ana","email":"ana@empresa.com","cargo":"Dev","departamento":"Ingeniería"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, esperado %d (CSRF ausente)", rr.Code, http.StatusForbidden)
	}
	if n := countEmpleados(t, a); n != 0 {
		t.Fatalf("se persistió estado con CSRF ausente: %d empleados", n)
	}
}

// Spec: "Invalid token is rejected" — token malformado: 403 y sin persistencia.
func TestCSRFMalformedRejectedWithoutStateChange(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withCookie(cookie),
		withHeader("X-CSRF-Token", "no-es-un-token-valido"),
		withBody(`{"nombre":"Ana","email":"ana@empresa.com","cargo":"Dev","departamento":"Ingeniería"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, esperado %d (CSRF malformado)", rr.Code, http.StatusForbidden)
	}
	if n := countEmpleados(t, a); n != 0 {
		t.Fatalf("se persistió estado con CSRF malformado: %d empleados", n)
	}
}

// Spec: "Invalid token is rejected" — token de OTRA sesión (desajustado):
// 403 y sin persistencia. Cada login emite un token de sesión único, por lo
// que el token A nunca valida contra la sesión B.
func TestCSRFMismatchedRejectedWithoutStateChange(t *testing.T) {
	a := newTestApp(t)
	cookieA := loginAs(t, a)
	cookieB := loginAs(t, a)
	tokenA := csrfFor(t, a, cookieA)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withCookie(cookieB),
		withHeader("X-CSRF-Token", tokenA),
		withBody(`{"nombre":"Ana","email":"ana@empresa.com","cargo":"Dev","departamento":"Ingeniería"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, esperado %d (CSRF de otra sesión)", rr.Code, http.StatusForbidden)
	}
	if n := countEmpleados(t, a); n != 0 {
		t.Fatalf("se persistió estado con CSRF desajustado: %d empleados", n)
	}
}

// Spec: "Valid token permits mutation" — operador autenticado con token válido
// (header X-CSRF-Token) puede ejecutar la mutación.
func TestValidCSRFTokenPermitsMutation(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	token := csrfFor(t, a, cookie)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withCookie(cookie),
		withHeader("X-CSRF-Token", token),
		withBody(`{"nombre":"Ana","email":"ana@empresa.com","cargo":"Dev","departamento":"Ingeniería"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, esperado %d (token válido)", rr.Code, http.StatusOK)
	}
	if n := countEmpleados(t, a); n != 1 {
		t.Fatalf("esperado 1 empleado persistido, hay %d", n)
	}
}

// Fallback: el token también se acepta como campo de formulario (_csrf) para
// requests urlencoded; con token válido la mutación procede.
func TestCSRFViaFormFieldAccepted(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	token := csrfFor(t, a, cookie)
	id := seedEmpleado(t, a)

	rr := a.request(t, http.MethodPost, "/api/recomendar/"+strconv.FormatInt(id, 10),
		withCookie(cookie),
		withBody("_csrf="+token),
		withContentType("application/x-www-form-urlencoded"))

	if rr.Code == http.StatusForbidden {
		t.Fatal("CSRF por campo de formulario rechazado con token válido")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, esperado %d (recomendar con CSRF por campo)", rr.Code, http.StatusOK)
	}
}
