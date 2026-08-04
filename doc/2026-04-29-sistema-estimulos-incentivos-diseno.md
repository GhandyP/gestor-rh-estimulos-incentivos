---
date: 2026-04-29
topic: "Sistema de Estímulos e Incentivos para RRHH"
status: validated
---

## Problem Statement

Recursos Humanos necesita una herramienta para mantener un **análisis descriptivo del capital humano** y generar **estímulos e incentivos personalizados** que mejoren el desempeño individual y organizacional.

El problema central: los sistemas tradicionales de incentivos aplican reglas fijas ("bono anual para todos igual") ignorando que cada persona tiene un **umbral de percepción distinto** y responde a **tipos de motivación diferentes**. Un estímulo mal calibrado es invisible; uno mal dirigido es ineficaz.

El sistema se fundamenta en dos modelos teóricos:

- **Modelo de estímulo y umbral (psicofísica)**: La intensidad necesaria para que un estímulo sea percibido varía según la persona y su punto de partida. Existen el umbral absoluto (mínimo detectable) y el umbral diferencial (cambio mínimo perceptible respecto a lo anterior).
- **Modelo de Fogg (B=MAP)**: Todo comportamiento requiere la convergencia simultánea de Motivación + Habilidad + Detonante (Prompt). Si alguno falta, la acción no ocurre.

## Constraints

- **MVP pequeño**: Alcance acotado a funcionalidad core, no una plataforma empresarial completa
- **Go puro**: Todo el backend en Go con dependencias mínimas
- **Stack simple**: SQLite sin CGO, HTML templates con HTMX, sin frameworks JS pesados
- **Sin integraciones externas en MVP**: Los datos de empleados se cargan manualmente (no se conecta a ERP/nómina)
- **Diseño portable**: La arquitectura debe permitir migrar SQLite → PostgreSQL sin reescribir lógica

## Approach

**Enfoque elegido: Sistema de tres capas conductuales paralelas (MAP) con umbrales calibrados por empleado.**

Cada capa ataca un componente del modelo de Fogg:

| Capa | Componente MAP | Naturaleza | Horizonte |
|------|---------------|------------|-----------|
| **Incentivos** | Motivación | Estratégica | Largo plazo / ciclos |
| **Nudges** | Habilidad | Ambiental | Siempre activo |
| **Estímulos** | Prompt (detonante) | Táctica | Inmediato / evento |

**Por qué tres capas y no solo una:** Un sistema que solo emite estímulos es "ruido" sin base motivacional. Un sistema con solo incentivos es invisible entre ciclos. La combinación crea un **campo conductual continuo** — el empleado siempre está dentro de un entorno diseñado (nudges), con motivadores visibles (incentivos), y recibe activaciones precisas en momentos óptimos (estímulos).

**Alternativas consideradas y descartadas:**
- CRUD tradicional con reglas fijas → No modela diferencias individuales, ignora la investigación conductual
- Plataforma de analytics con ML → Sobredimensionado para MVP, el modelo de Fogg es efectivo sin machine learning

## Architecture

### Visión general

```
                    B = M × A × P
                         │   │   │
    ┌────────────────────┼───┼───┼────────────────────┐
    │                    │   │   │                    │
    ▼                    ▼   ▼   ▼                    ▼
INCENTIVOS             NUDGES           ESTÍMULOS
(Estratégico)          (Ambiental)      (Táctico)
    │                    │                  │
    │  • Identidad       │  • Defaults      │  • Triggers calibrados
    │  • Beneficios      │  • Social proof  │  • Umbral personalizado
    │  • Formación       │  • Framing       │  • Timing óptimo
    │  • Proy.Corp       │  • Fricción 0    │  • Canal correcto
    │                    │                  │
    ▼                    ▼                  ▼
  Largo plazo         Siempre activo     Inmediato / evento
  Diseñado por RRHH   Diseño del entorno   Automatizado
  Grupos/roles        Toda la org         Individual
```

### Diagrama del sistema

```
┌──────────────────────────────────────────────────────────────┐
│                   CAPA DE PRESENTACIÓN                       │
│   Dashboard RRHH  │  API REST  │  Panel de Diseño de Nudges  │
├──────────────────────────────────────────────────────────────┤
│                   CAPA DE DOMINIO                            │
│                                                              │
│  ┌─────────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │  INCENTIVOS     │  │   NUDGES     │  │   ESTÍMULOS    │  │
│  │  (Motivación)   │  │  (Habilidad) │  │   (Prompt)     │  │
│  │                 │  │              │  │                │  │
│  │ • Identidad     │  │ • Defaults   │  │ • Calibración  │  │
│  │ • Beneficios    │  │ • Social     │  │ • Umbrales     │  │
│  │ • Formación     │  │ • Framing    │  │ • Timing       │  │
│  │ • Proy.Corp     │  │ • Fricción   │  │ • Canal        │  │
│  └────────┬────────┘  └──────┬───────┘  └───────┬────────┘  │
│           │                  │                   │            │
│           └──────────────────┼───────────────────┘            │
│                              ▼                               │
│  ┌───────────────────────────────────────────────────────┐   │
│  │              PERFIL MAP DEL EMPLEADO                   │   │
│  │  Motivación ← Sensible a qué tipo de incentivo        │   │
│  │  Habilidad  ← Dónde hay fricción / qué nudges aplicar │   │
│  │  Prompt     ← Umbral absoluto + diferencial           │   │
│  └───────────────────────────────────────────────────────┘   │
│                              │                               │
│  ┌───────────────────────────┴───────────────────────────┐   │
│  │           MOTOR DE ANÁLISIS DESCRIPTIVO                │   │
│  │  • Distribución MAP por depto/rol/antigüedad          │   │
│  │  • Efectividad por tipo de incentivo/nudge/estímulo   │   │
│  │  • Zona de riesgo (bajo la curva de acción)           │   │
│  │  • ROI de intervenciones                               │   │
│  └───────────────────────────────────────────────────────┘   │
├──────────────────────────────────────────────────────────────┤
│                   CAPA DE PERSISTENCIA                       │
│           SQLite (→ PostgreSQL para producción)              │
└──────────────────────────────────────────────────────────────┘
```

## Components

### 1. Perfil MAP del empleado

Cada empleado tiene tres puntuaciones (0.0–1.0), recalibradas periódicamente con datos observables:

- **Motivación**: Derivada de engagement, asistencia, feedback, antigüedad, historial de respuesta a incentivos previos
- **Habilidad / Capacidad**: Derivada de skills, certificaciones, brechas de competencia, carga de trabajo actual
- **Sensibilidad al Prompt**: Qué tan receptivo es a distintos tipos de detonantes (reconocimiento público, incentivo económico, tiempo libre, desarrollo profesional)

Las puntuaciones 0-1 permiten aplicar la "curva de acción" de Fogg directamente: si M × A < umbral del empleado, ningún prompt funcionará.

### 2. Capa de INCENTIVOS (Motivación — Estratégica)

Catálogo diseñado por RRHH. Los empleados los "alcanzan" por mérito o los reciben por criterio. Operan en ciclos (trimestrales, anuales).

| Categoría | Qué es | Ejemplos |
|-----------|--------|----------|
| **Identidad** | Refuerzan pertenencia y rol | Insignia "Senior Craftsmanship", embajador de cultura, mentor oficial |
| **Beneficios** | Tangibles más allá del salario | Días extra de vacaciones, presupuesto de home office, seguro ampliado |
| **Formación** | Crecimiento profesional | Beca para conferencia, acceso a plataforma, presupuesto de libros |
| **Proyectos Corporativos** | Participación en iniciativas estratégicas | Liderar un comité, representar a la empresa en evento, incubar idea propia |

Cada incentivo tiene: tipo, intensidad motivacional (0-1), costo para la empresa, criterios de elegibilidad, disponibilidad (cupos limitados, recurrente, permanente).

### 3. Capa de NUDGES (Habilidad — Ambiental)

Pequeños ajustes al entorno que reducen la fricción hacia comportamientos deseables. **No son notificaciones** — son diseño del entorno. Siempre activos.

| Tipo de Nudge | Mecanismo | Aplicación en RRHH |
|---------------|-----------|-------------------|
| **Defaults** | La opción por defecto es la deseable | Evaluación 360° pre-agendada, onboarding automático |
| **Social Proof** | Mostrar lo que otros hacen | "El 80% de tu equipo ya completó su plan de desarrollo" |
| **Framing** | Cómo se presenta la información | Enmarcar feedback como oportunidad, no como crítica |
| **Reducción de fricción** | Eliminar pasos innecesarios | One-click para solicitar capacitación, formularios auto-completados |

### 4. Capa de ESTÍMULOS (Prompt — Táctica)

El motor de triggers calibrados. Un estímulo solo se dispara si el empleado está sobre la curva de acción (M y A suficientes). Si no, el motor recomienda primero ajustar nudges o revisar elegibilidad de incentivos.

Para cada empleado, el sistema aprende:
- **Umbral absoluto**: Intensidad mínima del estímulo para que sea detectado como significativo
- **Umbral diferencial**: Cuánto debe incrementarse respecto al anterior para percibirse como mejora (Ley de Weber aplicada a RRHH). Ejemplo: si Juan recibió bono de $100, darle $105 está bajo su umbral diferencial (~constante de Weber 0.15 = ≥$115 para ser percibido).

### 5. Motor de recomendación

Orquesta las tres capas. Dado un objetivo de RRHH (ej: "mejorar retención en ingeniería"):

1. **Detector de riesgo**: Filtra empleados con M baja + A alta (riesgo de fuga)
2. **Capa incentivos**: ¿Hay incentivos que eleven M? Recomienda tipo + intensidad
3. **Capa nudges**: ¿Hay fricciones que reduzcan A? Recomienda ajuste ambiental
4. **Capa estímulos**: Solo cuando M y A son suficientes, dispara el prompt con timing y canal óptimos

### 6. Motor de análisis descriptivo

Paneles agregados que responden:
- Distribución de motivación por departamento
- Empleados en "zona de riesgo" (bajo la curva de acción)
- Efectividad histórica de tipos de incentivo/nudge/estímulo por perfil
- Costo de intervenciones vs. mejora de métricas
- Rotación predicha vs. rotación real

## Data Flow

```
1. RECOLECCIÓN (manual en MVP)
   RRHH ingresa: evaluaciones, métricas, feedback → Actualización perfil MAP

2. DETECCIÓN (automática)
   Motor monitorea curva de acción: ¿empleados bajo el umbral?
   Si sí → alerta a RRHH con diagnóstico (falta M, A, o P)

3. RECOMENDACIÓN (semi-automática)
   Motor orquesta las 3 capas:
   - Primero: ¿ajustar nudges? (más barato, menor fricción)
   - Segundo: ¿aplicar incentivo? (más costoso, mayor impacto)
   - Tercero: ¿disparar estímulo? (solo si M y A son suficientes)
   RRHH revisa, aprueba o ajusta

4. APLICACIÓN
   Se registra la intervención (tipo, destinatario, intensidad, fecha)

5. RETROALIMENTACIÓN (automática)
   Se mide respuesta: cambio en métricas post-intervención
   Se recalibra umbral del empleado
   El sistema ajusta recomendaciones futuras
```

### Flujo de decisión completo (ejemplo)

```
RRHH define objetivo: "Mejorar retención en ingeniería"

  ┌─→ MOTOR DE ANÁLISIS: ¿Quiénes están bajo la curva?
  │   Resultado: 3 empleados con M baja, A alta (riesgo fuga)
  │
  ├─→ CAPA INCENTIVOS: ¿Hay incentivos que eleven su M?
  │   Recomendación: "Proyecto Corporativo" para Juan
  │   (su perfil responde a autonomía + propósito)
  │
  ├─→ CAPA NUDGES: ¿Hay fricciones que bajen su A?
  │   Recomendación: Para María, simplificar proceso de
  │   solicitud de capacitación (quiere crecer pero el
  │   proceso actual es muy burocrático → nudges de fricción)
  │
  └─→ CAPA ESTÍMULOS: Una vez M y A están ok, disparar prompt
      Recomendación: Para Pedro (M y A ya óptimos), enviar
      invitación personalizada a liderar tech talk
      → timing: viernes AM, canal: email + Slack
```

## Error Handling

| Situación | Estrategia |
|-----------|------------|
| **Datos insuficientes para calibrar** | El sistema arranca con umbrales conservadores basados en defaults por rol/seniority. Afina con cada ciclo de retroalimentación |
| **Sobre-estimulación** | Si un empleado recibe estímulos muy frecuentes, el umbral absoluto sube (fatiga de incentivos). El sistema alerta cuando la frecuencia supera el threshold configurable |
| **Estímulo ineficaz** | Si un estímulo no produce cambio en métricas, el sistema no recalibra a ciegas — marca para revisión de RRHH (posible error en el perfil MAP subyacente) |
| **Sesgo en recomendaciones** | RRHH siempre tiene la última palabra. El sistema recomienda, no decide. Toda recomendación es trazable a los datos que la generaron |
| **Conflicto entre capas** | Si un nudge y un incentivo compiten por el mismo comportamiento, el motor prioriza: nudges primero (menor costo), incentivos después (mayor impacto). El orden es configurable |
| **Datos inconsistentes** | Validación en capa de dominio antes de persistir. Perfiles MAP requieren al menos N puntos de datos para ser considerados "confiables" |

## Testing Strategy

### Tests unitarios
- Motor de calibración de umbrales (absoluto + diferencial)
- Motor de recomendación (orquestación de 3 capas)
- Lógica de dominio MAP (cálculo de puntuaciones)
- Validación de reglas de elegibilidad de incentivos
- Cálculo de curva de acción de Fogg

### Tests de integración
- SQLite real con migraciones aplicadas
- Flujos completos: crear empleado → calibrar perfil → recomendar intervención → registrar resultado → recalibrar
- Transacciones y rollbacks

### Tests end-to-end
- API REST: crear/leer empleados, incentivos, nudges, estímulos
- Templates HTML: renderizado correcto del dashboard y panel de diseño de nudges

### Fixtures
- Set de empleados sintéticos con perfiles MAP conocidos:
  - "Juan": M=0.3, A=0.9, sensible a autonomía (debe gatillar alerta de riesgo fuga + recomendar incentivo de proyecto)
  - "María": M=0.7, A=0.2, sensible a formación (debe recomendar nudge de fricción antes que incentivo)
  - "Pedro": M=0.8, A=0.8 (debe ser candidato para estímulo directo)

## Stack Técnico

| Capa | Tecnología | Justificación |
|------|------------|---------------|
| Lenguaje | Go 1.22+ | Requerimiento explícito |
| HTTP Router | `net/http` (enhanced mux) | Stdlib, sin dependencias |
| Base de datos | SQLite vía `modernc.org/sqlite` | Sin CGO, portable, ideal para MVP |
| Migraciones | `golang-migrate` | Compatible con SQLite y PostgreSQL |
| Templates | `html/template` (stdlib) | Sin dependencias externas |
| Frontend | HTMX + CSS minimal | Interactividad sin JS framework |
| Validación | `go-playground/validator` | Estándar en ecosistema Go |
| Logging | `log/slog` (stdlib) | Structured logging nativo desde Go 1.21 |
| Testing | `testing` + `testify` | Assertions y mocking |

## Estructura del Proyecto

```
estimulos-incentivos/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── domain/                     # Entidades puras (sin dependencias externas)
│   │   ├── empleado.go
│   │   ├── perfil_map.go
│   │   ├── incentivo.go            # Tipos: identidad, beneficio, formación, proyecto_corp
│   │   ├── nudge.go                # Tipos: defaults, social_proof, framing, friccion
│   │   ├── estimulo.go             # Trigger calibrado con intensidad + canal + timing
│   │   └── umbral.go               # Umbral absoluto + diferencial + histórico
│   ├── engine/                     # Lógica de negocio pura
│   │   ├── calibrador.go           # Cálculo y ajuste de umbrales
│   │   ├── recomendador.go         # Orquestación de las 3 capas
│   │   ├── detector_riesgo.go      # Detección de empleados bajo la curva de acción
│   │   └── analisis.go             # Motor descriptivo (distribuciones, efectividad, ROI)
│   ├── store/                      # Capa de persistencia
│   │   └── sqlite/
│   │       ├── empleados.go
│   │       ├── incentivos.go
│   │       ├── nudges.go
│   │       ├── estimulos.go
│   │       ├── historial.go        # Registro de intervenciones y respuestas
│   │       └── migrations/
│   ├── handler/                    # HTTP handlers
│   │   ├── empleados.go
│   │   ├── incentivos.go
│   │   ├── nudges.go
│   │   ├── estimulos.go
│   │   ├── recomendaciones.go
│   │   └── dashboard.go
│   └── service/                    # Capa de aplicación (orquesta domain + engine + store)
│       ├── empleado_service.go
│       ├── recomendacion_service.go
│       └── analisis_service.go
├── web/
│   └── templates/                  # Templates Go html/template
│       ├── base.html
│       ├── dashboard.html
│       ├── empleados/
│       ├── incentivos/
│       ├── nudges/
│       └── estimulos/
├── go.mod
└── go.sum
```

## Open Questions

1. **Origen de datos en MVP**: ¿Los datos de empleados y métricas se ingresan 100% manual por formulario, o hay algún CSV/Excel que se importe? Esto define si necesitamos un importador en el MVP.
2. **Métricas de desempeño**: ¿Qué métricas concretas vamos a trackear? (ej: cumplimiento de objetivos, asistencia, feedback 360°, productividad). Esto define la estructura del perfil MAP.
3. **Alcance del dashboard**: ¿El MVP necesita gráficos (distribución MAP, efectividad) o alcanza con tablas y alertas textuales?
4. **Multi-tenant**: ¿Una instancia = una empresa, o el MVP contempla múltiples organizaciones?
