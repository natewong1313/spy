CREATE TABLE company (
    name VARCHAR(255) NOT NULL PRIMARY KEY,
    platform_type VARCHAR(255) NOT NULL,
    platform_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    greenhouse_name VARCHAR(255)
);

CREATE TABLE job (
    url TEXT NOT NULL PRIMARY KEY,
    company VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    locations VARCHAR(255)[] NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_company FOREIGN KEY (company) REFERENCES company(name) ON DELETE CASCADE
);



