-- 用户表（认证）
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_login_at INTEGER
);

-- 索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);