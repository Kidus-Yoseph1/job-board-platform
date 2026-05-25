CREATE TABLE jobs (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  title        TEXT        NOT NULL,
  description  TEXT        NOT NULL,
  category     TEXT        NOT NULL,
  location     TEXT        NOT NULL,
  type         TEXT        NOT NULL
               CHECK (type IN ('full_time', 'part_time', 'contract', 'remote')),
  is_negotiable BOOLEAN    NOT NULL DEFAULT true,
  status       TEXT        NOT NULL DEFAULT 'draft'
               CHECK (status IN ('draft', 'open', 'closed')),
  deleted_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON jobs (company_id);
CREATE INDEX ON jobs (category) WHERE deleted_at IS NULL;
CREATE INDEX ON jobs (location) WHERE deleted_at IS NULL;
CREATE INDEX ON jobs (status)   WHERE deleted_at IS NULL;
CREATE INDEX ON jobs (type)     WHERE deleted_at IS NULL;