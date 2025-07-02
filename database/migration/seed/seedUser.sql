-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
INSERT INTO users (name, email, password, created_by, updated_by)
VALUES
  ('Admin', 'admin@example.com', 'hashed_password_placeholder', 'system', 'system');