-- +goose Up
CREATE TABLE IF NOT EXISTS movies (
    id              BIGINT      PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    release_year    INT         NOT NULL,
    director        VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS movies;