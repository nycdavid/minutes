CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    application TEXT NOT NULL,
    start_timestamp INTEGER NOT NULL,
    end_timestamp INTEGER NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
