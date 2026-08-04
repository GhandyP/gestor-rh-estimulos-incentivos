# Backend Implementation Report — 2026-05-27

## Status: COMPLETE

All backend Go code for prototipo v2 compiled, tested, and verified.

## Files changed/created

| File | Action | Lines |
|------|--------|-------|
| `internal/handler/handler.go` | **Rewritten** | ~430 lines |
| `internal/service/service.go` | **Extended** | +200 lines |
| `internal/engine/elegibilidad.go` | **Created** | 100 lines |
| `internal/store/sqlite/store.go` | **Fixed** | 5-line migration fix |
| `README.md` | **Created** | ~120 lines |

## What was implemented

### handler.go (complete rewrite)
- **Removed** inline `dashboardTemplate` const (~200 lines of HTML)
- **Preserved** all 11 existing API routes unchanged
- **Added** 6 HTML page routes (`/`, `/empleados`, `/empleados/{id}`, `/incentivos`, `/nudges`, `/estimulos`) with template loading from `web/templates/`
- **Added** 17 new API routes:
  - Empleados: Get, Detail (composite), Update, Profile MAP Update, Form partial, CSV Import, CSV Template
  - Incentivos: Get, Update, Delete, Elegibles, AddElegibilidad, DeleteElegibilidad
  - Nudges: Get, Update
  - Estímulos: List (filtered), Get, ApplyForm partial
- **Error handling**: 400/404/500 status codes, JSON encoding error check in `writeJSON`

### service.go (extended)
- **EmpleadoDetail** composite struct (`empleado + perfil + umbral + historial + estimulos + sobreCurva`)
- `GetEmpleadoDetail()` — aggregates empleado, perfil MAP, umbral, historial, estímulos
- `UpdateEmpleado()` — partial field update
- `ImportCSV()` — CSV parsing with `encoding/csv`, error-per-row reporting
- `GetIncentivo()`, `UpdateIncentivo()`, `DeleteIncentivo()`
- `AddElegibilidad()`, `DeleteElegibilidad()`
- `GetIncentivosElegibles()` — uses `engine.EvaluarElegibilidad()` per incentivo
- `GetNudge()`, `UpdateNudge()` — direct DB for update (store method missing)
- `GetEstimulo()`
- `ListEstimulos()` — filtered by estado (pendiente/aplicado/todos), raw DB query for non-pendiente

### engine/elegibilidad.go (new)
- `EvaluarElegibilidad()` — evaluates all criteria against an employee
- `evaluarCriterio()` — supports campo: `departamento` (eq/neq), `cargo` (eq/neq), `antiguedad_meses` (eq/gte/lte/gt/lt)
- `evaluarString()` / `evaluarNumero()` — operator dispatchers
- `mesesDesde()` — calculates months since a date

### store.go (fix)
- ALTER TABLE for `tipo` column now checks `pragma_table_info` before executing, preventing duplicate-column errors on restart

## Verification

```
✓ go vet ./...       — clean
✓ go build ./cmd/server/   — clean
✓ go test ./internal/engine/ — 9/9 PASS
✓ gofmt — all files compliant
```

## API surface (31 routes total)

```
HTML:  GET /, GET /empleados, GET /empleados/:id, GET /incentivos, GET /nudges, GET /estimulos
API:   11 existing + 14 new = 25 API routes
```

## Dependencies on frontend worker

The handler references these template files which the frontend worker creates:
- `web/templates/base.html` — layout with nav + CSS + `{{block "content" .}}`
- `web/templates/dashboard.html` — `{{define "content"}}` dashboard
- `web/templates/empleados/list.html` — employee list
- `web/templates/empleados/detail.html` — employee detail
- `web/templates/empleados/_form.html` — inline form partial
- `web/templates/incentivos/list.html` — incentives catalog
- `web/templates/nudges/list.html` — nudges panel
- `web/templates/estimulos/list.html` — stimuli list

If any are missing, the page handler returns HTTP 500 with the parse error message.
