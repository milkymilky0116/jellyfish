-- +goose Up
CREATE TABLE email_category (
  email_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY (email_id, category_id),
  FOREIGN KEY (email_id) REFERENCES email (id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose Down
DROP TABLE email_category;

