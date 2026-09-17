package main

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"estimulos-incentivos/internal/service"
	"estimulos-incentivos/internal/store/sqlite"
)

// Spec (portfolio-documentation): las cuatro escenarios de documentación
// (deploy seguro, límites del threat model, demo consistente con la
// implementación y non-goals explícitos) requieren evidencia runtime. Estos
// tests son pruebas de contrato de documentación: leen README.md y
// doc/design-production-hardening.md (y ejecutan Service.Seed sobre una base
// SQLite real en el escenario de demo) y verifican que la documentación
// cumple los contratos de las escenarios. RED: no existía ninguna prueba de
// contrato; GREEN: los contenidos reales cumplen cada contrato (estilo
// approval: se fija el contrato para detectar regresiones de documentación).
// No usan red, Docker ni servicios externos.

// mustSeedStore crea una base SQLite real en un directorio temporal, aplica
// las migraciones embebidas y ejecuta Service.Seed (el mismo camino de
// arranque de producción). La base se cierra automáticamente al terminar.
func mustSeedStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "demo.db")
	st, err := sqlite.New(dsn)
	if err != nil {
		t.Fatalf("sqlite.New(%s): %v", dsn, err)
	}
	t.Cleanup(func() { _ = st.Close() })

	svc := service.New(st)
	if err := svc.Seed(context.Background()); err != nil {
		t.Fatalf("Service.Seed: %v", err)
	}
	return st
}

// countRows cuenta las filas de una tabla en la base real sembrada.
func countRows(t *testing.T, st *sqlite.Store, table string) int {
	t.Helper()
	var n int
	if err := st.DB().QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// Scenario: "New operator can deploy safely" — un lector de la documentación
// debe poder identificar persistencia, salud, supuestos de seguridad y pasos
// de verificación para el camino local y de contenedor.
func TestDocsContractNewOperatorCanDeploySafely(t *testing.T) {
	readme := readRepoFile(t, "README.md")
	design := readRepoFile(t, "doc/design-production-hardening.md")

	categories := []struct {
		name   string
		needle string
		where  string // "readme" | "design" | "any"
	}{
		{"deployment local", "go run ./cmd/server", "readme"},
		{"deployment container", "docker compose up --build", "readme"},
		{"persistence path", "DB_PATH", "any"},
		{"persistence volume", "estimulos-data:/data", "any"},
		{"persistence backup guidance", "back up", "any"},
		{"health liveness", "/healthz", "any"},
		{"health readiness", "/readyz", "any"},
		{"security operator credentials", "OPERATOR_USER", "any"},
		{"security session secret", "SESSION_SECRET", "any"},
		{"verification test", "go test ./...", "any"},
		{"verification build", "go build ./...", "any"},
		{"verification vet", "go vet ./...", "any"},
		{"verification format gate", "gofmt -l .", "any"},
	}

	for _, c := range categories {
		t.Run(c.name, func(t *testing.T) {
			switch {
			case c.where == "readme":
				if !strings.Contains(readme, c.needle) {
					t.Errorf("README.md debe contener %q para el contrato 'New operator can deploy safely'", c.needle)
				}
			case c.where == "design":
				if !strings.Contains(design, c.needle) {
					t.Errorf("design debe contener %q para el contrato 'New operator can deploy safely'", c.needle)
				}
			default: // "any": basta con que uno de los dos documentos lo contenga
				if !strings.Contains(readme, c.needle) && !strings.Contains(design, c.needle) {
					t.Errorf("README.md o el design doc deben contener %q para el contrato 'New operator can deploy safely'", c.needle)
				}
			}
		})
	}
}

// normalizeMarkdown quita marcadores de énfasis de negrita/cursiva para que
// las búsquedas de frases sean robustas ante el formato (p. ej. "**one**
// trusted operator" → "one trusted operator").
func normalizeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "`", "")
	return s
}

// Scenario: "Threat model states boundaries" — el alcance de un solo operador,
// los supuestos de host confiable, las mutaciones protegidas y las funciones
// empresariales excluidas deben estar explícitos.
func TestDocsContractThreatModelStatesBoundaries(t *testing.T) {
	readme := normalizeMarkdown(readRepoFile(t, "README.md"))
	design := normalizeMarkdown(readRepoFile(t, "doc/design-production-hardening.md"))

	boundaries := []struct {
		name   string
		needle string
	}{
		{"single-operator scope", "single-operator"},
		{"one trusted operator", "one trusted operator"},
		{"trusted host assumption", "trusted"},
		{"protected mutations require CSRF", "CSRF"},
		{"mutations require session and token", "session"},
		{"excluded SSO", "SSO"},
		{"excluded multi-tenancy", "multi-tenant"},
		{"excluded multi-role RBAC", "RBAC"},
		{"excluded external integrations", "external integrations"},
	}

	for _, b := range boundaries {
		t.Run(b.name, func(t *testing.T) {
			if !strings.Contains(readme, b.needle) && !strings.Contains(design, b.needle) {
				t.Errorf("README o design doc deben declarar el límite %q del threat model", b.needle)
			}
		})
	}
}

// Scenario: "Demo matches implementation" — los datos de demo documentados y
// el comportamiento del listado de empleados deben coincidir con la
// implementación real. Prueba runtime: ejecuta Service.Seed sobre una base
// SQLite real (t.TempDir) y compara los conteos reales con los conteos
// documentados en el README; además verifica que el README describe el
// comportamiento ya cubierto por los tests de render del listado.
func TestDocsContractDemoMatchesImplementation(t *testing.T) {
	readme := readRepoFile(t, "README.md")

	// Claims documentados del demo, p. ej. "(6 employees, 8 incentives, 6 nudges)".
	re := regexp.MustCompile(`\((\d+) employees, (\d+) incentives, (\d+) nudges\)`)
	m := re.FindStringSubmatch(readme)
	if m == nil {
		t.Fatal("README.md debe documentar los conteos del demo: '(N employees, N incentives, N nudges)'")
	}
	docCounts := []int{atoiOrFail(t, m[1]), atoiOrFail(t, m[2]), atoiOrFail(t, m[3])}

	// Implementación real: Seed determinista sobre una base nueva.
	st := mustSeedStore(t)
	realCounts := []int{
		countRows(t, st, "empleados"),
		countRows(t, st, "incentivos"),
		countRows(t, st, "nudges"),
	}

	names := []string{"employees", "incentives", "nudges"}
	for i, want := range docCounts {
		if realCounts[i] != want {
			t.Errorf("demo documentado (%d %s) no coincide con Service.Seed (%d %s reales)", want, names[i], realCounts[i], names[i])
		}
	}

	// El README debe describir el comportamiento ya probado del listado:
	// exactamente una acción de borrado por fila (cubierto en
	// TestEmpleadosListExactlyOneDeleteActionPerRow) sin reclamar más.
	if !strings.Contains(readme, "exactly one") || !strings.Contains(readme, "hx-delete") {
		t.Error("README.md debe documentar el comportamiento probado del listado (exactamente una acción de borrado por fila vía hx-delete)")
	}
}

// Scenario: "Non-goals remain explicit" — todas las capacidades excluidas
// deben identificarse claramente como fuera de alcance en el README o el
// design doc (SSO/OIDC, multi-tenancy, RBAC multi-rol, integraciones
// externas, jobs, PostgreSQL, Kubernetes, ML/analytics y browser E2E).
func TestDocsContractNonGoalsRemainExplicit(t *testing.T) {
	readme := readRepoFile(t, "README.md")
	design := readRepoFile(t, "doc/design-production-hardening.md")

	nonGoals := []struct {
		name   string
		needle string
	}{
		{"no SSO/OIDC", "SSO"},
		{"no multi-tenancy", "multi-tenant"},
		{"no multi-role RBAC", "RBAC"},
		{"no external integrations", "external integrations"},
		{"no background jobs", "jobs"},
		{"no PostgreSQL", "PostgreSQL"},
		{"no Kubernetes", "Kubernetes"},
		{"no analytics/ML", "ML"},
		{"no browser E2E", "browser E2E"},
	}

	for _, ng := range nonGoals {
		t.Run(ng.name, func(t *testing.T) {
			if !strings.Contains(readme, ng.needle) && !strings.Contains(design, ng.needle) {
				t.Errorf("README o design doc deben declarar el non-goal %q", ng.needle)
			}
		})
	}

	// Los non-goals deben estar enmarcados como excluidos (no como capacidades).
	for _, marker := range []string{"non-goal", "non-goals", "Out of scope", "out of scope"} {
		if strings.Contains(readme, marker) {
			return
		}
	}
	t.Error("README.md debe enmarcar las exclusiones como non-goals (non-goals / out of scope)")
}

func atoiOrFail(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("Atoi(%q): %v", s, err)
	}
	return n
}
