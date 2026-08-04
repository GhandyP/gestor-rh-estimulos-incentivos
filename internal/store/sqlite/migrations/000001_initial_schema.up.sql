CREATE TABLE empleados (
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

CREATE TABLE perfiles_map (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id),
    motivacion REAL NOT NULL DEFAULT 0.5,
    habilidad REAL NOT NULL DEFAULT 0.5,
    prompt REAL NOT NULL DEFAULT 0.5,
    sensibilidad TEXT NOT NULL DEFAULT 'desarrollo',
    confiabilidad REAL NOT NULL DEFAULT 0.3,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE incentivos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    descripcion TEXT NOT NULL,
    tipo TEXT NOT NULL CHECK(tipo IN ('identidad', 'beneficios', 'formacion', 'proyecto_corporativo')),
    intensidad REAL NOT NULL DEFAULT 0.5,
    costo REAL NOT NULL DEFAULT 0,
    disponibilidad TEXT NOT NULL DEFAULT 'permanente' CHECK(disponibilidad IN ('limitado', 'recurrente', 'permanente')),
    cupos INTEGER NOT NULL DEFAULT 0,
    cupos_usados INTEGER NOT NULL DEFAULT 0,
    activo INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE elegibilidades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    incentivo_id INTEGER NOT NULL REFERENCES incentivos(id),
    campo TEXT NOT NULL,
    operador TEXT NOT NULL,
    valor TEXT NOT NULL
);

CREATE TABLE nudges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    descripcion TEXT NOT NULL,
    tipo TEXT NOT NULL CHECK(tipo IN ('defaults', 'social_proof', 'framing', 'friccion')),
    ambito TEXT NOT NULL DEFAULT 'global' CHECK(ambito IN ('global', 'departamento', 'individual')),
    target_id INTEGER NOT NULL DEFAULT 0,
    activo INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE estimulos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL REFERENCES empleados(id),
    tipo TEXT NOT NULL DEFAULT '',
    contenido TEXT NOT NULL DEFAULT '',
    intensidad REAL NOT NULL DEFAULT 0.5,
    canal TEXT NOT NULL DEFAULT 'email' CHECK(canal IN ('email', 'slack', 'presencial', 'dashboard')),
    estado TEXT NOT NULL DEFAULT 'pendiente' CHECK(estado IN ('pendiente', 'aplicado', 'fallido')),
    fecha_ideal TEXT NOT NULL DEFAULT (datetime('now')),
    fecha_aplicado TEXT,
    origen_recomendacion TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE umbrales (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empleado_id INTEGER NOT NULL UNIQUE REFERENCES empleados(id),
    umbral_absoluto REAL NOT NULL DEFAULT 0.3,
    umbral_diferencial REAL NOT NULL DEFAULT 0.15,
    ultimo_estimulo REAL NOT NULL DEFAULT 0,
    fecha_ultimo_estimulo TEXT,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE historial_estimulos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    umbral_id INTEGER NOT NULL REFERENCES umbrales(id),
    fecha TEXT NOT NULL DEFAULT (datetime('now')),
    intensidad REAL NOT NULL,
    respuesta_map REAL NOT NULL DEFAULT 0,
    tipo TEXT NOT NULL DEFAULT ''
);
