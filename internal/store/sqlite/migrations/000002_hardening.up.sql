-- 000002_hardening.up.sql
-- Hardening de integridad: limpieza defensiva de huérfanos, reconstrucción de
-- tablas con CHECK y FKs explícitas (ON DELETE CASCADE) e índices de acceso.
--
-- PRAGMA foreign_keys es una no-op dentro de una transacción; la reconstrucción
-- corre en autocommit con FKs desactivadas durante el proceso y se reactivan al
-- final. La validación final (PRAGMA foreign_key_check) la ejecuta el runner.

PRAGMA foreign_keys = OFF;

-- ---------------------------------------------------------------------------
-- Backfill / validación de filas existentes
-- Las filas hijas sin padre son datos irrecuperables: se eliminan antes de
-- activar las restricciones para que la migración sea segura sobre bases
-- heredadas (p. ej. borrados previos con FKs desactivadas).
-- ---------------------------------------------------------------------------
DELETE FROM historial_estimulos
 WHERE umbral_id NOT IN (SELECT id FROM umbrales);

DELETE FROM perfiles_map
 WHERE empleado_id NOT IN (SELECT id FROM empleados);

DELETE FROM umbrales
 WHERE empleado_id NOT IN (SELECT id FROM empleados);

DELETE FROM estimulos
 WHERE empleado_id NOT IN (SELECT id FROM empleados);

DELETE FROM elegibilidades
 WHERE incentivo_id NOT IN (SELECT id FROM incentivos);

-- ---------------------------------------------------------------------------
-- Reconstrucción: empleados
-- ---------------------------------------------------------------------------
CREATE TABLE empleados_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    cargo TEXT NOT NULL,
    departamento TEXT NOT NULL,
    fecha_ingreso TEXT NOT NULL,
    activo INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    CHECK (length(trim(email)) > 0),
    CHECK (length(trim(nombre)) > 0)
);

INSERT INTO empleados_new (id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at)
SELECT id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at FROM empleados;

DROP TABLE empleados;
ALTER TABLE empleados_new RENAME TO empleados;

-- ---------------------------------------------------------------------------
-- Reconstrucción: perfiles_map
-- ---------------------------------------------------------------------------
CREATE TABLE perfiles_map_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id) ON DELETE CASCADE,
    motivacion REAL NOT NULL DEFAULT 0.5 CHECK (motivacion BETWEEN 0 AND 1),
    habilidad REAL NOT NULL DEFAULT 0.5 CHECK (habilidad BETWEEN 0 AND 1),
    prompt REAL NOT NULL DEFAULT 0.5 CHECK (prompt BETWEEN 0 AND 1),
    sensibilidad TEXT NOT NULL DEFAULT 'desarrollo'
        CHECK (sensibilidad IN ('economico', 'reconocimiento', 'desarrollo', 'bienestar')),
    confiabilidad REAL NOT NULL DEFAULT 0.3 CHECK (confiabilidad BETWEEN 0 AND 1),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO perfiles_map_new (id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
SELECT id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at FROM perfiles_map;

DROP TABLE perfiles_map;
ALTER TABLE perfiles_map_new RENAME TO perfiles_map;

-- ---------------------------------------------------------------------------
-- Reconstrucción: estimulos
-- ---------------------------------------------------------------------------
CREATE TABLE estimulos_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL REFERENCES empleados(id) ON DELETE CASCADE,
    tipo TEXT NOT NULL DEFAULT '',
    contenido TEXT NOT NULL DEFAULT '',
    intensidad REAL NOT NULL DEFAULT 0.5 CHECK (intensidad BETWEEN 0 AND 1),
    canal TEXT NOT NULL DEFAULT 'email' CHECK (canal IN ('email', 'slack', 'presencial', 'dashboard')),
    estado TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'aplicado', 'fallido')),
    fecha_ideal TEXT NOT NULL DEFAULT (datetime('now')),
    fecha_aplicado TEXT,
    origen_recomendacion TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO estimulos_new (id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at)
SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at FROM estimulos;

DROP TABLE estimulos;
ALTER TABLE estimulos_new RENAME TO estimulos;

-- ---------------------------------------------------------------------------
-- Reconstrucción: umbrales
-- ---------------------------------------------------------------------------
CREATE TABLE umbrales_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id) ON DELETE CASCADE,
    umbral_absoluto REAL NOT NULL DEFAULT 0.3 CHECK (umbral_absoluto BETWEEN 0 AND 1),
    umbral_diferencial REAL NOT NULL DEFAULT 0.15 CHECK (umbral_diferencial BETWEEN 0 AND 1),
    ultimo_estimulo REAL NOT NULL DEFAULT 0 CHECK (ultimo_estimulo BETWEEN 0 AND 1),
    fecha_ultimo_estimulo TEXT,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO umbrales_new (id, empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at)
SELECT id, empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at FROM umbrales;

DROP TABLE umbrales;
ALTER TABLE umbrales_new RENAME TO umbrales;

-- ---------------------------------------------------------------------------
-- Reconstrucción: historial_estimulos
-- ---------------------------------------------------------------------------
CREATE TABLE historial_estimulos_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    umbral_id INTEGER NOT NULL REFERENCES umbrales(id) ON DELETE CASCADE,
    fecha TEXT NOT NULL DEFAULT (datetime('now')),
    intensidad REAL NOT NULL CHECK (intensidad BETWEEN 0 AND 1),
    respuesta_map REAL NOT NULL DEFAULT 0 CHECK (respuesta_map BETWEEN 0 AND 1),
    tipo TEXT NOT NULL DEFAULT ''
);

INSERT INTO historial_estimulos_new (id, umbral_id, fecha, intensidad, respuesta_map, tipo)
SELECT id, umbral_id, fecha, intensidad, respuesta_map, tipo FROM historial_estimulos;

DROP TABLE historial_estimulos;
ALTER TABLE historial_estimulos_new RENAME TO historial_estimulos;

-- ---------------------------------------------------------------------------
-- Índices de acceso para las consultas del store y el servicio.
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_historial_umbral_fecha ON historial_estimulos(umbral_id, fecha);
CREATE INDEX IF NOT EXISTS idx_estimulos_estado_fecha_ideal ON estimulos(estado, fecha_ideal);
CREATE INDEX IF NOT EXISTS idx_estimulos_empleado_estado ON estimulos(empleado_id, estado);

PRAGMA foreign_keys = ON;
