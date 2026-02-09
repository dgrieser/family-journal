CREATE TABLE IF NOT EXISTS session_store (
    id VARCHAR(255) PRIMARY KEY,
    data BLOB NOT NULL,
    expires_at DATETIME NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_session_store_expires_at (expires_at)
);
