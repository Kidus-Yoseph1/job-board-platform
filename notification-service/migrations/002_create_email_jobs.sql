CREATE TABLE email_jobs (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  to_email   TEXT        NOT NULL,
  subject    TEXT        NOT NULL,
  body       TEXT        NOT NULL,
  status     TEXT        NOT NULL DEFAULT 'pending'
             CHECK (status IN ('pending', 'sent', 'failed')),
  attempts   INT         NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON email_jobs (status) WHERE status = 'pending';