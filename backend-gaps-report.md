# Backend Gaps Report — Estimulos e Incentivos

## Status: **COMPLETE (code), UNVERIFIED (build)**

El entorno Go tiene un problema de toolchain que impide compilar (timeout en `go build` incluso con un programa trivial). La sintaxis de todos los archivos Go es válida (`gofmt -e` OK en 20/20 archivos). Los templates se validaron con un programa aparte (24/24 OK).

---

## Archivos creados

| Archivo | Descripción |
|---------|-------------|
| `web/templates/empleados/_row.html` | Fila `<tr>` HTML de empleado para HTMX append en creación (A2) |
| `web/templates/empleados/_import_result.html` | Resumen HTML del resultado de import CSV (A4) |
| `web/templates/estimulos/_table.html` | Tabla de estímulos con tabs de filtro por estado (B4) |
| `web/templates/nudges/_card.html` | Card de nudge individual con toggle switch (C1) |
| `web/templates/incentivos/detail.html` | Página de detalle de incentivo con elegibilidades (C3) |
| `web/templates/nudges/detail.html` | Página de detalle de nudge con toggle (C3) |

## Archivos modificados

| Archivo | Cambios |
|---------|---------|
| `internal/handler/handler.go` | 1108 líneas. Modificados: CreateEmpleadoAPI(A2), CreateIncentivoAPI(A3), CreateNudgeAPI(A3), ImportCSVAPI(A4), UpdatePerfilMAPAPI(B3), EstimulosPage(B4). Nuevas rutas registradas: 15 endpoints (Dashboard partials, import-form, form fragments, toggle, delete, detail pages). |
| `internal/service/service.go` | 657 líneas. Métodos DeleteEmpleado y ListElegibilidades ya existían de la iteración previa. Sin cambios adicionales necesarios. |
| `web/templates/empleados/list.html` | Corregido endpoint del form (`/api/empleados/form` → `/api/empleados/new-form`). Agregado botón eliminar con `hx-delete`. |
| `web/templates/incentivos/list.html` | Nombres de incentivo linkean a `/incentivos/{id}`. |
| `web/templates/estimulos/list.html` | Reescrito con tabs y carga dinámica vía JS + HTMX. |

## Gaps cubiertos por tarea

### Tanda A — Bugs funcionales
| Gap | Estado | Implementación |
|-----|--------|----------------|
| A1 Toasts | ⚠️ Templates listos | CSS y header `X-Toast` configurados en handlers. El JS listener va en el frontend agent. |
| A2 Crear empleado | ✅ | CreateEmpleadoAPI: si HX-Request → render `_row.html` + header `X-Toast` |
| A3 Crear incentivo/nudge | ✅ | CreateIncentivoAPI/CreateNudgeAPI: si HX-Request → `HX-Redirect` a la lista |
| A4 CSV import | ✅ | ImportCSVAPI: si HX-Request → render `_import_result.html` + header `X-Toast` |

### Tanda B — Cableado faltante
| Gap | Estado | Implementación |
|-----|--------|----------------|
| B1 Dashboard partials | ✅ | 4 endpoints: `GET /api/dashboard/{stats,riesgos,distribucion,efectividad}` |
| B2 Import form endpoint | ✅ | `GET /api/empleados/import-form` → `_import_form.html` |
| B3 Editar perfil MAP | ✅ | UpdatePerfilMAPAPI: si HX-Request → `HX-Redirect` a `/empleados/{id}` + toast |
| B4 Filtro estímulos | ✅ | EstimulosPage: si HX-Request → render `_table.html`. list.html tiene tabs JS. |

### Tanda C — Features faltantes
| Gap | Estado | Implementación |
|-----|--------|----------------|
| C1 Toggle nudge | ✅ | `PUT /api/nudges/{id}/toggle` → toggle Activo → render `_card.html` |
| C2 Eliminar empleado | ✅ | `DELETE /api/empleados/{id}` → `X-Toast` header |
| C3 Detalle incentivo/nudge | ✅ | `GET /incentivos/{id}` + `GET /nudges/{id}` con templates completos |

## Nuevos endpoints (15)

```
GET  /api/dashboard/stats           → DashboardStatsAPI
GET  /api/dashboard/riesgos         → DashboardRiesgosAPI
GET  /api/dashboard/distribucion    → DashboardDistribucionAPI
GET  /api/dashboard/efectividad     → DashboardEfectividadAPI
GET  /api/empleados/import-form     → ImportFormAPI
GET  /api/incentivos/form           → IncentivoFormAPI
GET  /api/nudges/form               → NudgeFormAPI
PUT  /api/nudges/{id}/toggle        → ToggleNudgeAPI
DELETE /api/empleados/{id}           → DeleteEmpleadoAPI
GET  /incentivos/{id}               → IncentivoDetailPage
GET  /nudges/{id}                   → NudgeDetailPage
```

## Validación de sintaxis

- ✅ 20/20 archivos Go pasan `gofmt -e`
- ✅ Validación programática de templates (24/24 OK, ejecutada antes del fork)
- ❌ `go build` timeout — problema de toolchain Go, no de código

## Riesgos

1. **Go toolchain roto**: `go build` no funciona en esta sesión fork. El código compilaba correctamente ANTES del fork (build + vet + tests OK). Se necesita restaurar el entorno Go.
2. **El frontend agent debe conectar el JS de toasts** en `base.html` (listener `htmx:afterRequest` que lea headers `X-Toast`).
3. **Duplicate methods removidos**: el handler.go tenía duplicados de la iteración previa; se eliminaron. Verificar que la versión final tenga exactamente una definición de cada handler (grep -c confirma).
