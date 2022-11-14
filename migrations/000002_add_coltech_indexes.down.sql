-- Filename: migrations/000002_add_coltech_indexes.down.sql

DROP INDEX IF EXISTS tblcoltech_created_by_idx;
DROP INDEX IF EXISTS tblcoltech_assigned_to_idx;
DROP INDEX IF EXISTS tblcoltech_priority_val_idx;
DROP INDEX IF EXISTS tblcoltech_status_val_idx;