CREATE TABLE authors (
    id BIGSERIAL PRIMARY KEY,
    name text NOT NULL,
    bio text
);  

CREATE INDEX IF NOT EXISTS "idx_authors_id" ON "authors" ("id");