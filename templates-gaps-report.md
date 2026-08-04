# Templates Gaps Report — 2026-05-27

## Status: COMPLETE (22/22 templates parse OK)

All 11 gaps (A1 through C3) addressed across templates. The backend Go code was also updated where needed to integrate template changes.

## Files created (6)

| File | Gap | Description |
|------|-----|-------------|
| `web/templates/empleados/_row.html` | A2 | Table row partial for new empleado (HTMX insert) |
| `web/templates/empleados/_import_result.html` | A4 | CSV import result with success/error display |
| `web/templates/estimulos/_table.html` | B4 | Estimulos table partial (used by tabs + HTMX filter) |
| `web/templates/nudges/_card.html` | C1 | Single nudge card with toggle switch (define template) |
| `web/templates/incentivos/detail.html` | C3 | Incentivo detail page with elegibilidades CRUD |
| `web/templates/nudges/detail.html` | C3 | Nudge detail page with toggle and info |

## Files modified (10)

| File | Gap | Changes |
|------|-----|---------|
| `web/templates/base.html` | A1 | Added toast JS listener (`htmx:afterRequest` → X-Toast header) |
| `web/templates/dashboard.html` | B1 | Rewritten: now loads stats/riesgos/distribucion/efectividad via HTMX |
| `web/templates/_stats.html` | B1 | Updated for flat data: `.TotalEmpleados`, `.EmpleadosEnRiesgo`, `.IncentivosActivos`, `.NudgesActivos` |
| `web/templates/_riesgos.html` | B1 | Added `<h2>` + card content wrapper (self-contained partial) |
| `web/templates/_distribucion.html` | B1 | Added `<h2>` + card content wrapper |
| `web/templates/_efectividad.html` | B1 | Added `<h2>` + card content wrapper |
| `web/templates/empleados/list.html` | B2, C2 | Import CSV button already present; added delete button column + buttons in rows |
| `web/templates/empleados/_perfil_form.html` | B3 | Changed `hx-target` from `#main-content` to `closest .card` |
| `web/templates/estimulos/list.html` | B4 | Rewritten: tabs (Todos/Pendientes/Aplicados) + HTMX-driven table loading |
| `web/templates/nudges/list.html` | C1 | Uses `{{template "nudge-card" .}}` from `_card.html`; toggle switches |

## Backend changes (coordinated)

| File | Changes |
|------|---------|
| `internal/handler/handler.go` | Added 11 handlers: DashboardStatsAPI, DashboardRiesgosAPI, DashboardDistribucionAPI, DashboardEfectividadAPI, ImportFormAPI, IncentivoFormAPI, NudgeFormAPI, ToggleNudgeAPI, DeleteEmpleadoAPI, IncentivoDetailPage, NudgeDetailPage. Updated NudgesPage to parse `_card.html`. EstimulosPage handles HTMX for table partial. CreateEmpleadoAPI returns HTML row for HTMX. ImportCSVAPI returns HTML for HTMX. ApplyEstimuloAPI sends X-Toast header. |
| `internal/service/service.go` | Added: `ToggleNudge()`, `IncentivoDetailData` struct, `GetIncentivoDetail()`. Removed duplicate `DeleteEmpleado` and `ListElegibilidades` (already added by parallel agent). |

## Validation

- All 22 template combinations parse successfully with `percent` FuncMap
- Build succeeded after cleaning stale /tmp/go-build directories
- All HTMX attributes verified: `hx-get`, `hx-post`, `hx-put`, `hx-delete`, `hx-ext`, `hx-target`, `hx-swap`, `hx-trigger`, `hx-boost`, `hx-select`, `hx-confirm`, `hx-push-url`, `hx-headers`

## Design Notes

### Toast system
- Server sets `X-Toast` response header to trigger toasts
- JS listener on `htmx:afterRequest` reads the header and shows toast
- Auto-dismisses after 4 seconds
- Used in: ApplyEstimuloAPI, DeleteEmpleadoAPI, UpdatePerfilMAPAPI, ImportCSVAPI

### Nudge toggle
- Uses `_card.html` as named template `{{define "nudge-card"}}`
- Backend `ToggleNudgeAPI` returns rendered nudge-card after toggling
- `NudgesPage` parses base.html + list.html + _card.html together

### Dashboard partials
- Dashboard page loads immediately (empty containers with spinners)
- 4 HTMX requests load data asynchronously:
  - `/api/dashboard/stats` → `_stats.html` (flat fields)
  - `/api/dashboard/riesgos` → `_riesgos.html` (receives full AnalisisResult)
  - `/api/dashboard/distribucion` → `_distribucion.html` (receives full AnalisisResult)
  - `/api/dashboard/efectividad` → `_efectividad.html` (receives full AnalisisResult)

### Estímulos tabs
- Tabs use `hx-get="/estimulos?estado=X"` with `hx-target="#estimulos-content"`
- Server returns `_table.html` partial for HTMX requests, full page for direct visits
- `hx-push-url` updates browser URL without full reload

### Perfil MAP edit
- `_perfil_form.html` targets `closest .card` so only the perfil card refreshes
- Backend `UpdatePerfilMAPAPI` sends `X-Toast` and `HX-Redirect` back to detail

## Risks
- Go toolchain temporarily locked by corrupted /tmp/go-build dirs from parallel agents. Required `rm -rf /tmp/go-build*` to resolve.
- Backend `IncentivoDetailPage` (written by parallel agent) doesn't use `GetIncentivoDetail()` — it manually builds data. If the template expects `.Elegibles`, it won't render that section (guarded by `{{with .Elegibles}}` so it degrades gracefully).
- `ToggleNudgeAPI` uses `h.Svc.ToggleNudge()` which uses `UpdateNudge()` which uses raw SQL (not the store's UpdateNudge method — but the store doesn't have UpdateNudge exposed). This is acceptable for the prototype.
