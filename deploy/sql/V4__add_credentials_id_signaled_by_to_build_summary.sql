ALTER TABLE credentials ADD UNIQUE (id);
ALTER TABLE build_summary ADD COLUMN signaled_by SMALLINT;
/*set all signaled_by to poll signal for now*/
UPDATE build_summary SET signaled_by = 2;
ALTER TABLE build_summary ALTER COLUMN signaled_by SET NOT NULL;

ALTER TABLE build_summary ADD COLUMN credentials_id INT;
UPDATE build_summary SET (credentials_id) =
                           (
                             SELECT id FROM credentials
                             WHERE credentials.account = build_summary.account AND credentials.cred_type=1
                           )
;

ALTER TABLE build_summary ALTER COLUMN credentials_id SET NOT NULL;

