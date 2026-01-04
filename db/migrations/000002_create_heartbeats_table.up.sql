CREATE TABLE heartbeats (
    id INTEGER PRIMARY KEY,
    application TEXT NOT NULL,
    timestamp INTEGER NOT NULL,

    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)
