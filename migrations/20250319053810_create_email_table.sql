-- +goose Up
CREATE TABLE email (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  seq INTEGER NOT NULL,
  sender TEXT NOT NULL,
  subject TEXT NOT NULL,
  email_date DATETIME NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE email;

