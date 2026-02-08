CREATE TABLE IF NOT EXISTS sessions (
    k VARCHAR(255) NOT NULL,
    v LONGBLOB NOT NULL,
    e BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (k),
    KEY idx_sessions_expiration (e)
);
