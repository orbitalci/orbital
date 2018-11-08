ALTER TABLE credentials ADD COLUMN id SERIAL;
ALTER TABLE polling_repos ADD COLUMN credentials_id INT;

UPDATE polling_repos SET (credentials_id) =
  (
    SELECT id FROM credentials
    WHERE credentials.cred_type = 1 AND credentials.account = polling_repos.account
  )
;

ALTER TABLE polling_repos ALTER COLUMN credentials_id SET NOT NULL;
