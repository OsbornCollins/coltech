-- Filename: migrations/000001_create_coltech_list_table.up.sql

CREATE TABLE IF NOT EXISTS tblcoltech (
    id bigserial PRIMARY KEY,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    summary text NOT NULL,
    description text NOT NULL,
    priority_val text NOT NULL DEFAULT 'MEDIUM',
    status_val text NOT NULL DEFAULT 'OPEN',
    assigned_to text NOT NULL DEFAULT 'UNASSIGNED',
    category text NOT NULL,
    department text NOT NULL,
    closed_on timestamp(0) with time zone DEFAULT '0001-01-01 00:00:00',
    due_on timestamp(0) with time zone DEFAULT '0001-01-01 00:00:00',
    created_by text NOT NULL,
    version int NOT NULL DEFAULT 1
);