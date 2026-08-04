# RRHH · Sistema de Estímulos e Incentivos

Sistema de análisis de capital humano basado en el modelo conductual de **Fogg (B = M × A × P)**.

Diseñado para RRHH: mantener un análisis descriptivo del capital humano y generar estímulos e incentivos personalizados que mejoren el desempeño individual y organizacional.

## Quick start

```bash
go run ./cmd/server/
# Abrir http://localhost:8080
```

Variables de entorno opcionales:

- `PORT` — puerto HTTP (default: `8080`)
- `DB_PATH` — ruta a la base SQLite (default: `estimulos.db`)

Al primer inicio se crea la base de datos y se cargan datos de demo (6 empleados, 8 incentivos, 6 nudges).

## Stack técnico

| Capa | Tecnología |
|------|-----------|
| Lenguaje | Go 1.26 |
| HTTP | `net/http` (enhanced mux Go 1.22+) |
| Base de datos | SQLite via `modernc.org/sqlite` (sin CGO) |
| Templates | `html/template` stdlib |
| Frontend | HTMX 2.0 + CSS vanilla |
| Testing | `testing` stdlib |

## Modelo teórico

El sistema se fundamenta en el modelo de comportamiento de **BJ Fogg**:

```
B = M × A × P
```

Todo comportamiento (B) requiere la convergencia simultánea de:

- **M**otivación — el deseo de actuar
- **H**abilidad — la capacidad de actuar
- **P**rompt — el detonante que activa la acción

Si alguno de los tres falta, el comportamiento no ocurre.

### Tres capas conductuales

| Capa | Componente MAP | Naturaleza | Horizonte |
|------|---------------|------------|-----------|
| **Incentivos** | Motivación | Estratégica | Largo plazo / ciclos |
| **Nudges** | Habilidad | Ambiental | Siempre activo |
| **Estímulos** | Prompt (detonante) | Táctica | Inmediato / evento |

Cada empleado tiene un **perfil MAP** (puntuaciones 0.0–1.0) y un **umbral psicofísico** que determina la intensidad mínima necesaria para que un estímulo sea percibido.

## API

### HTML (navegación)

| Ruta | Descripción |
|------|-------------|
| `GET /` | Dashboard con análisis y métricas |
| `GET /empleados` | Lista de empleados |
| `GET /empleados/{id}` | Detalle de empleado con perfil MAP |
| `GET /incentivos` | Catálogo de incentivos |
| `GET /nudges` | Panel de nudges ambientales |
| `GET /estimulos` | Estímulos (filtro `?estado=pendiente`) |

### API REST

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/empleados` | Listar empleados |
| `POST` | `/api/empleados` | Crear empleado |
| `GET` | `/api/empleados/{id}` | Obtener empleado |
| `GET` | `/api/empleados/{id}/detail` | Detalle compuesto (JSON) |
| `PUT` | `/api/empleados/{id}` | Actualizar empleado |
| `PUT` | `/api/empleados/{id}/perfil` | Actualizar perfil MAP |
| `POST` | `/api/empleados/import` | Importar CSV |
| `GET` | `/api/empleados/template` | Descargar plantilla CSV |
| `GET` | `/api/analisis` | Análisis descriptivo |
| `GET` | `/api/riesgos` | Zona de riesgo |
| `POST` | `/api/recomendar/{id}` | Recomendar intervención |
| `GET` | `/api/incentivos` | Listar incentivos |
| `POST` | `/api/incentivos` | Crear incentivo |
| `GET` | `/api/incentivos/{id}` | Obtener incentivo |
| `PUT` | `/api/incentivos/{id}` | Actualizar incentivo |
| `DELETE` | `/api/incentivos/{id}` | Eliminar incentivo |
| `GET` | `/api/incentivos/elegibles/{empleadoId}` | Incentivos elegibles |
| `POST` | `/api/incentivos/{id}/elegibilidades` | Agregar criterio |
| `DELETE` | `/api/elegibilidades/{id}` | Eliminar criterio |
| `GET` | `/api/nudges` | Listar nudges |
| `POST` | `/api/nudges` | Crear nudge |
| `GET` | `/api/nudges/{id}` | Obtener nudge |
| `PUT` | `/api/nudges/{id}` | Actualizar nudge |
| `GET` | `/api/estimulos` | Listar estímulos |
| `GET` | `/api/estimulos/{id}` | Obtener estímulo |
| `POST` | `/api/estimulos/{id}/aplicar` | Aplicar estímulo |

## Arquitectura

```
cmd/server/main.go            → Entry point
internal/
  domain/                     → Entidades puras (sin dependencias externas)
  engine/                     → Lógica de negocio (calibrador, recomendador, detector, analisis, elegibilidad)
  store/sqlite/               → Persistencia con SQLite + migraciones embebidas
  handler/                    → HTTP handlers (REST API + vistas HTML)
  service/                    → Orquestación dominio/engine/store
web/templates/                → Templates Go html/template + HTMX partials
```

## Flujo principal

1. **Recolección**: RRHH ingresa empleados (manual o CSV)
2. **Detección**: El motor monitorea la curva de acción Fogg
3. **Recomendación**: Orquesta las 3 capas (incentivos → nudges → estímulos)
4. **Aplicación**: Se registra la intervención
5. **Retroalimentación**: Se mide respuesta y se recalibran umbrales

## Licencia

Privado — MVP para uso interno.
