-- name: CreateCategory :one
INSERT INTO category (name, key, modseq) VALUES (?, ?, ?) RETURNING *;

-- name: GetCategory :one
SELECT * FROM category WHERE key = ? LIMIT 1;
