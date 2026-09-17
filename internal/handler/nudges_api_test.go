package handler

import (
	"net/http"
	"strings"
	"testing"
)

// TestCreateNudgeAPI_ErrorInternoNoFiltrado: un error del store en la
// creación de nudges responde un mensaje genérico, sin detalles internos
// (constraint SQL, driver, etc.).
func TestCreateNudgeAPI_ErrorInternoNoFiltrado(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	csrf := a.auth.CSRFToken(&Session{Token: cookie.Value})

	// ambito fuera del CHECK de la BD: garantiza un error del store.
	rr := a.request(t, http.MethodPost, "/api/nudges",
		withCookie(cookie),
		withHeader("X-CSRF-Token", csrf),
		withContentType("application/json"),
		withBody(`{"nombre":"Prueba","tipo":"defaults","ambito":"invalido"}`))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado %d", rr.Code, http.StatusInternalServerError)
	}
	body := rr.Body.String()
	for _, leaked := range []string{"SQL", "constraint", "CHECK", "sqlite"} {
		if strings.Contains(body, leaked) {
			t.Errorf("la respuesta 500 filtra detalle interno: %q en %q", leaked, body)
		}
	}
	if !strings.Contains(body, "error interno") {
		t.Errorf("la respuesta 500 debe decir 'error interno', obtuvo %q", body)
	}
}
