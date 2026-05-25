CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE notifications (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    TEXT        NOT NULL,
  type       TEXT        NOT NULL
             CHECK (type IN ('job_applied', 'application_status_changed', 'job_status_changed')),
  title      TEXT        NOT NULL,
  body       TEXT        NOT NULL,
  read       BOOLEAN     NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON notifications (user_id);
CREATE INDEX ON notifications (user_id, read) WHERE read = false;