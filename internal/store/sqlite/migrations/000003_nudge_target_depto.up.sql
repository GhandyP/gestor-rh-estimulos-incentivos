-- 000003_nudge_target_depto.up.sql
-- Los departamentos son TEXT en empleados.departamento; el objetivo de un
-- nudge de departamento necesita una columna de texto propia. target_id
-- (INTEGER) queda reservado para el ámbito individual.
ALTER TABLE nudges ADD COLUMN target_depto TEXT NOT NULL DEFAULT '';
