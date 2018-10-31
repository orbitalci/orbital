ALTER TABLE credentials ADD COLUMN id SERIAL;
ALTER TABLE polling_repos ADD COLUMN credentials_id INT UNIQUE;

UPDATE polling_repos SET (credentials_id) =
  (
    SELECT id FROM credentials
    WHERE credentials.account = polling_repos.account
  )
;