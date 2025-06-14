CREATE TABLE IF NOT EXISTS alerts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    severity TEXT,
    source TEXT,
    acked_by TEXT,
    acked_at TIMESTAMP,
    code TEXT,
    count INTEGER,
    author TEXT
);