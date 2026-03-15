CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    oauth_provider TEXT,
    oauth_id TEXT UNIQUE,
    email TEXT,
    username TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);