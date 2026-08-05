package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// --- Rutas de éxito (JSON API, sin template rendering) -----------------------

// Spec: "Authenticated operator is allowed" — operador autenticado alcanza
// las rutas de lectura del API con 200 y JSON. Listas vacías y recursos con
// datos reales (seed via service, igual que el flujo real).
func TestAuthenticatedSuccessRoutes(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	id := seedEmpleado(t, a)

	cases := []struct {
		name string
		path string
		want string // subcadena esperada en el body ("" = solo status+JSON)
	}{
		{"list-empleados", "/api/empleados", "ana@empresa.com"},
		{"get-empleado", "/api/empleados/" + strconv.FormatInt(id, 10), "Ana"},
		{"list-incentivos", "/api/incentivos", ""},
		{"list-estimulos", "/api/estimulos", ""},
		{"list-nudges", "/api/nudges", ""},
		{"analisis", "/api/analisis", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := a.request(t, http.MethodGet, tc.path, withCookie(cookie))
			if rr.Code != http.StatusOK {
				t.Fatalf("%s: status %d, want 200", tc.path, rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("%s: Content-Type %q, want JSON", tc.path, ct)
			}
			if tc.want != "" && !strings.Contains(rr.Body.String(), tc.want) {
				t.Errorf("%s: body no contiene %q: %s", tc.path, tc.want, rr.Body.String())
			}
		})
	}
}

// --- Rutas de error (mapeo seguro, sin datos internos) -----------------------

// Spec: "safe 4xx/5xx errors" — IDs inválidos, recursos ausentes y métodos
// no permitidos devuelven su status esperado y el body NO expone stack traces
// ni pánicos internos. (Nota: un GET a una ruta desconocida cae en el patrón
// "GET /" del dashboard — comportamiento preexistente; el 405 cubre el
// rechazo por método en rutas no mapeadas.)
func TestErrorRoutesReturnSafeStatuses(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	token := csrfFor(t, a, cookie)

	cases := []struct {
		name, method, path string
		want               int
	}{
		{"method-not-allowed", http.MethodPost, "/api/no-existe", http.StatusMethodNotAllowed},
		{"invalid-empleado-id", http.MethodGet, "/api/empleados/abc", http.StatusBadRequest},
		{"missing-empleado", http.MethodGet, "/api/empleados/99999", http.StatusNotFound},
		{"invalid-estimulo-id", http.MethodGet, "/api/estimulos/abc", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := []reqOpt{withCookie(cookie)}
			if tc.method != http.MethodGet {
				// Las mutaciones exigen CSRF válido para llegar al mux.
				opts = append(opts, withHeader("X-CSRF-Token", token))
			}
			rr := a.request(t, tc.method, tc.path, opts...)
			if rr.Code != tc.want {
				t.Fatalf("%s: status %d, want %d", tc.path, rr.Code, tc.want)
			}
			body := rr.Body.String()
			if strings.Contains(body, "goroutine") || strings.Contains(body, "panic") {
				t.Errorf("%s: body expone detalles internos: %s", tc.path, body)
			}
		})
	}
}

// Spec: "HTTP boundary rejects unsafe requests" — un body malformado con
// autenticación Y CSRF válidos no persiste nada: 400 sin estado parcial.
func TestMalformedJSONBodyRejectedWithoutStateChange(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	token := csrfFor(t, a, cookie)

	rr := a.request(t, http.MethodPost, "/api/empleados",
		withCookie(cookie),
		withHeader("X-CSRF-Token", token),
		withBody(`{"nombre":`),
		withContentType("application/json"))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want %d (JSON malformado)", rr.Code, http.StatusBadRequest)
	}
	if n := countEmpleados(t, a); n != 0 {
		t.Fatalf("empleados = %d, want 0 (body malformado no persiste)", n)
	}
}

// Spec: "Invalid token is rejected ... changes no data" — la persistencia
// intacta también aplica a un SEGUNDO verbo de mutación: PUT sin CSRF se
// rechaza en el middleware y el empleado queda sin cambios.
func TestRejectedUpdateMutationDoesNotChangeState(t *testing.T) {
	a := newTestApp(t)
	cookie := loginAs(t, a)
	id := seedEmpleado(t, a)

	rr := a.request(t, http.MethodPut, "/api/empleados/"+strconv.FormatInt(id, 10),
		withCookie(cookie),
		withBody(`{"nombre":"Renombrado"}`),
		withContentType("application/json"))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, want %d (PUT sin CSRF)", rr.Code, http.StatusForbidden)
	}

	got, err := a.h.Svc.GetEmpleado(context.Background(), id)
	if err != nil || got == nil {
		t.Fatalf("get empleado: %v, %v", got, err)
	}
	if got.Nombre != "Ana" {
		t.Fatalf("nombre = %q, want %q (sin cambios tras PUT rechazado)", got.Nombre, "Ana")
	}
}
