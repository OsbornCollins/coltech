-- Filename: migrations/000002_add_coltech_indexes.up.sql

CREATE INDEX IF NOT EXISTS tblcoltech_created_by_idx ON tblcoltech USING GIN(to_tsvector('simple', created_by));
CREATE INDEX IF NOT EXISTS tblcoltech_assigned_to_idx ON tblcoltech USING GIN(to_tsvector('simple', assigned_to));
CREATE INDEX IF NOT EXISTS tblcoltech_priority_idx ON tblcoltech USING GIN(to_tsvector('simple', priority));
CREATE INDEX IF NOT EXISTS tblcoltech_status_idx ON tblcoltech USING GIN(to_tsvector('simple', status));