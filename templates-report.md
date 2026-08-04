# Templates Report — Estimulos e Incentivos

## Summary
18 template files created/rewritten for the RRHH system. All templates use Go html/template syntax, extend base layout, support HTMX interactivity, and follow the existing CSS design system.

## Files Created (18)

### Layout & Dashboard
| File | Type | Description |
|------|------|-------------|
| `base.html` | Layout | Base HTML with `<head>`, full CSS (400+ lines), nav bar, toast container, HTMX CDN, main-content slot |
| `dashboard.html` | Page | 4 KPI stats, MAP distribution bars by depto, risk zone table, effectiveness chart |
| `_stats.html` | Partial | 4 stat cards (empleados, riesgo, incentivos, nudges) |
| `_riesgos.html` | Partial | Risk zone table with severity badges |
| `_distribucion.html` | Partial | MAP bars per department (M, A, P) |
| `_efectividad.html` | Partial | Effectiveness table with success rate bars |

### Empleados
| File | Type | Description |
|------|------|-------------|
| `empleados/list.html` | Page | Employee table + "Nuevo Empleado" + "Importar CSV" buttons |
| `empleados/_form.html` | Form | Inline create form (nombre, email, cargo, departamento) |
| `empleados/_import_form.html` | Form | CSV file upload form with template download link |
| `empleados/detail.html` | Page | Full detail: personal info, MAP profile with editable bars, psychometric threshold, stimulus history |
| `empleados/_perfil_form.html` | Form | Edit MAP (M, A, P, sensitivity, confidence) |

### Incentivos
| File | Type | Description |
|------|------|-------------|
| `incentivos/list.html` | Page | Card grid of incentives with type badges, intensity, cost, availability |
| `incentivos/_form.html` | Form | Create incentive form (name, type, description, intensity, cost, availability) |

### Nudges
| File | Type | Description |
|------|------|-------------|
| `nudges/list.html` | Page | Grouped panel by type (Defaults, Social Proof, Framing, Fricción) |
| `nudges/_form.html` | Form | Create nudge form (name, type, description, scope, target) |

### Estímulos
| File | Type | Description |
|------|------|-------------|
| `estimulos/list.html` | Page | Table of stimuli with status and "Aplicar" action |
| `estimulos/_apply_form.html` | Form | Apply stimulus form with response input |

### Recommendation
| File | Type | Description |
|------|------|-------------|
| `_recomendacion.html` | Partial | Recommendation result card with diagnosis, incentive/nudge recs, generated stimulus |

## CSS Design System
Preserved from existing dashboard and extended with:
- **New components:** Toast notifications, HTMX spinners, form styles, toggle switches, tabs, nudge cards, chart bars, empty states, Fogg curve indicators
- **New badge variants:** badge-identidad, badge-beneficios, badge-formacion, badge-proyecto, badge-defaults, badge-social, badge-framing, badge-friccion, badge-info
- **New utilities:** .toolbar, .form-row, .form-inline, .detail-grid, .detail-item, .detail-label, .detail-value, .nudge-group, .nudge-card
- **Responsive:** Media query at 768px for mobile

## Go Backend Requirements
The Go handler must register this FuncMap:
```go
"percent": func(v float64) string { return fmt.Sprintf("%.0f", v*100) }
```

New routes needed:
- `GET /empleados`, `GET /incentivos`, `GET /nudges`, `GET /estimulos` (page renders)
- `GET /empleados/{id}` (detail page)
- `GET /api/empleados/form`, `GET /api/empleados/{id}/perfil-form` (form fragments)
- `GET /api/incentivos/form`, `GET /api/nudges/form` (form fragments)
- `GET /api/estimulos/{id}/apply-form` (apply form fragment)
- `GET /api/empleados/import-form` (import form fragment)
- `POST /api/empleados/import` (multipart CSV upload)
- `PUT /api/empleados/{id}/perfil` (MAP profile update)
- `POST /api/nudges/{id}/toggle` (nudge toggle)
- `GET /api/empleados/template` (CSV template download)
- `GET /api/estimulos` (list stimuli)

Data structures passed to templates must include fields matching domain types: `.Empleados`, `.Incentivos`, `.Nudges`, `.Estimulos`, `.Analisis`, `.Empleado`, `.Perfil`, `.Umbral`, `.Historial`.

## Validation
- All 18 files verified present in `web/templates/`
- All templates use valid Go html/template syntax (`{{define}}`, `{{block}}`, `{{template}}`, `{{range}}`, `{{if}}`, `{{with}}`)
- All pages extend `base.html` via `{{template "base" .}}` with `{{define "content"}}`
- Partials are HTML fragments (no base extension)
- All HTMX attributes correct: `hx-get`, `hx-post`, `hx-put`, `hx-target`, `hx-swap`, `hx-boost`, `hx-select`, `hx-encoding`, `hx-indicator`
- Spanish language throughout

## Risks
- `percent` FuncMap must be registered on Go side — without it, bar widths and percentage displays will be empty
- Date method calls (`.Format`) depend on Go template support for method invocation on struct fields
- `.IntensidadMinimaPerceptible` method call in detail template requires no pointer issues
- The `hx-select="#main-content"` on nav links may need adjustment based on actual DOM structure
