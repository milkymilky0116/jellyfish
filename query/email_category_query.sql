-- name: RegisterEmailAndCategory :exec
INSERT INTO email_category (email_id, category_id) VALUES (?, ?);
