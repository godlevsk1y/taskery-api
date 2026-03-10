CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_owner_id_title
ON tasks (owner_id, title);