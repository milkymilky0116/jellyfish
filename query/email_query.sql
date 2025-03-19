-- name: GetEmailById :one
SELECT * FROM email WHERE id = ? LIMIT 1;

-- name: CreateEmail :one
INSERT INTO email (seq, sender, subject, email_date) VALUES (?, ?, ?, ?) RETURNING *;
