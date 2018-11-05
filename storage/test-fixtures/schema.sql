-- noinspection SqlNoDataSourceInspectionForFile

CREATE TABLE build_summary (
  hash character varying(50),
  failed boolean default true,
  starttime timestamp without time zone DEFAULT '1970-01-01 00:00:00',
  account character varying(100),
  buildtime numeric default -99.999,
  repo character varying(100),
  id SERIAL PRIMARY KEY,
  branch character varying(100),
  queuetime timestamp without time zone DEFAULT '1970-01-01 00:00:00',
  status integer
);

CREATE TABLE build_output (
  build_id BIGINT,
  output bytea,
  id SERIAL PRIMARY KEY,
  FOREIGN KEY (build_id) REFERENCES build_summary (id) ON DELETE CASCADE
);

CREATE TABLE build_stage_details (
  id SERIAL PRIMARY KEY,
  stage text,
  build_id BIGINT,
  error text,
  starttime  timestamp without time zone,
  runtime numeric,
  status integer,
  messages jsonb,
  FOREIGN KEY (build_id) REFERENCES build_summary (id) ON DELETE CASCADE
);

create table polling_repos (
  account character varying(100),
  repo character varying(100),
  cron_string character varying (50),
  last_cron_time timestamp without time zone,
  branches character varying,
  last_hashes jsonb default '{}',
  primary key (account, repo)
);


CREATE TABLE credentials (
  account character varying(100),
  identifier character varying(100),
  cred_type smallint,
  cred_sub_type smallint,
  additional_fields jsonb,
  primary key (account, identifier, cred_sub_type)
);

ALTER TABLE credentials ADD COLUMN id SERIAL;
ALTER TABLE polling_repos ADD COLUMN credentials_id INT UNIQUE;

UPDATE polling_repos SET (credentials_id) =
(
  SELECT id FROM credentials
  WHERE credentials.account = polling_repos.account
)
;

ALTER TABLE polling_repos ALTER COLUMN credentials_id SET NOT NULL;