CREATE TABLE submissions (
    submission_id TEXT PRIMARY KEY,
    user_id TEXT,
    problem_id TEXT,
    code TEXT,
    language TEXT,
    verdict TEXT,
    execution_time BIGINT,
    memory_used BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    message TEXT,
    priority TEXT
);