CREATE TABLE build_summary (
  hash character varying(50),
  failed boolean,
  starttime timestamp without time zone,
  account character varying(100),
  buildtime numeric,
  repo character varying(100),
  id SERIAL PRIMARY KEY,
  branch character varying(100)
);

CREATE TABLE build_output (
  build_id BIGINT,
  output bytea,
  id SERIAL PRIMARY KEY,

  FOREIGN KEY (build_id) REFERENCES build_summary (id) ON DELETE CASCADE
);

CREATE TABLE build_failure_reason (
  build_id BIGINT,
  reasons jsonb,
  id SERIAL PRIMARY KEY,

  FOREIGN KEY (build_id) REFERENCES build_summary (id) ON DELETE CASCADE
);