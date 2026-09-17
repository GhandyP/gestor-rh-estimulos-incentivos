-- 000003_nudge_target_depto.down.sql
-- Rollback: elimina solo la columna agregada en 000003 (SQLite >= 3.35
-- soporta DROP COLUMN; el bundle modernc.org/sqlite la incluye).
ALTER TABLE nudges DROP COLUMN target_depto;
