CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL CHECK ( length(trim(username)) > 0 ),
    email TEXT NOT NULL CHECK ( length(trim(email)) > 0 ),
    password_hash TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users(lower(email));

CREATE TABLE IF NOT EXISTS tasks(
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    title TEXT NOT NULL CHECK ( length(trim(title)) > 0 ),
    description TEXT NOT NULL DEFAULT '',

    deadline TIMESTAMP WITH TIME ZONE NULL,

    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_owner_id ON tasks(owner_id);
