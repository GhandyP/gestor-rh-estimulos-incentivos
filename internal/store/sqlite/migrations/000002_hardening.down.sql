-- 000002_hardening.down.sql
-- Rollback que preserva datos: elimina solo lo agregado en 000002
-- (índices, CHECKs y ON DELETE CASCADE) devolviendo las tablas a la forma
-- de 000001 sin perder ninguna fila. schema_migrations queda intacto; el
-- runner no reaplica 000002 porque su versión ya está registrada.
--
-- Ejecutar con FKs desactivadas durante la reconstrucción, igual que en el up.

PRAGMA foreign_keys = OFF;

DROP INDEX IF EXISTS idx_historial_umbral_fecha;
DROP INDEX IF EXISTS idx_estimulos_estado_fecha_ideal;
DROP INDEX IF EXISTS idx_estimulos_empleado_estado;

CREATE TABLE empleados_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    cargo TEXT NOT NULL,
    departamento TEXT NOT NULL,
    fecha_ingreso TEXT NOT NULL,
    activo INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO empleados_old (id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at)
SELECT id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at FROM empleados;

DROP TABLE empleados;
ALTER TABLE empleados_old RENAME TO empleados;

CREATE TABLE perfiles_map_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id),
    motivacion REAL NOT NULL DEFAULT 0.5,
    habilidad REAL NOT NULL DEFAULT 0.5,
    prompt REAL NOT NULL DEFAULT 0.5,
    sensibilidad TEXT NOT NULL DEFAULT 'desarrollo',
    confiabilidad REAL NOT NULL DEFAULT 0.3,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO perfiles_map_old (id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
SELECT id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at FROM perfiles_map;

DROP TABLE perfiles_map;
ALTER TABLE perfiles_map_old RENAME TO perfiles_map;

CREATE TABLE estimulos_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL REFERENCES empleados(id),
    tipo TEXT NOT NULL DEFAULT '',
    contenido TEXT NOT NULL DEFAULT '',
    intensidad REAL NOT NULL DEFAULT 0.5,
    canal TEXT NOT NULL DEFAULT 'email' CHECK (canal IN ('email', 'slack', 'presencial', 'dashboard')),
    estado TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'aplicado', 'fallido')),
    fecha_ideal TEXT NOT NULL DEFAULT (datetime('now')),
    fecha_aplicado TEXT,
    origen_recomendacion TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO estimulos_old (id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at)
SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at FROM estimulos;

DROP TABLE estimulos;
ALTER TABLE estimulos_old RENAME TO estimulos;

CREATE TABLE umbrales_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id),
    umbral_absoluto REAL NOT NULL DEFAULT 0.3,
    umbral_diferencial REAL NOT NULL DEFAULT 0.15,
    ultimo_estimulo REAL NOT NULL DEFAULT 0,
    fecha_ultimo_estimulo TEXT,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO umbrales_old (id, empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at)
SELECT id, empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at FROM umbrales;

DROP TABLE umbrales;
ALTER TABLE umbrales_old RENAME TO umbrales;

CREATE TABLE historial_estimulos_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    umbral_id INTEGER NOT NULL REFERENCES umbrales(id),
    fecha TEXT NOT NULL DEFAULT (datetime('now')),
    intensidad REAL NOT NULL,
    respuesta_map REAL NOT NULL DEFAULT 0,
    tipo TEXT NOT NULL DEFAULT ''
);

INSERT INTO historial_estimulos_old (id, umbral_id, fecha, intensidad, respuesta_map, tipo)
SELECT id, umbral_id, fecha, intensidad, respuesta_map, tipo FROM historial_estimulos;

DROP TABLE historial_estimulos;
ALTER TABLE historial_estimulos_old RENAME TO historial_estimulos;

PRAGMA foreign_keys = ON;
