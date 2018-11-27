CREATE TABLE active_subscriptions (
  subscribed_to_acct_repo character varying(500),
  subscribing_acct_repo character varying(500),
  branch_queue_map jsonb,
  insert_time timestamp without time zone DEFAULT '1970-01-01 00:00:00',
  id SERIAL PRIMARY KEY
);

CREATE TABLE subscription_data (
  build_id INTEGER REFERENCES build_summary(id) ON DELETE CASCADE,
  active_subscriptions_id INTEGER REFERENCES active_subscriptions(id),
  subscribed_to_build_id INTEGER REFERENCES build_summary(id)
);
