CREATE TABLE job (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    company VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    locations VARCHAR(255)[] NOT NULL,
    url TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
