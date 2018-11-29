CREATE TABLE active_subscriptions (
  subscribed_to_vcs_cred_type SMALLINT, /*subcredtype for the upstream acct/repo (ie bitbucket)*/
  subscribed_to_repo character varying(150),
  subscribing_vcs_cred_type SMALLINT, /*subcredtype for the downstream acct/repo (ie bitbucket)*/
  subscribing_repo character varying(150),
  branch_queue_map jsonb,
  insert_time timestamp without time zone DEFAULT '1970-01-01 00:00:00',
  alias character varying(25),
  id SERIAL UNIQUE,
  PRIMARY KEY (subscribed_to_repo, subscribed_to_vcs_cred_type, subscribing_repo, subscribing_vcs_cred_type)
);

CREATE TABLE subscription_data (
  build_id INTEGER REFERENCES build_summary(id) ON DELETE CASCADE,
  active_subscriptions_id INTEGER REFERENCES active_subscriptions(id),
  subscribed_to_build_id INTEGER REFERENCES build_summary(id)
);
