# project ocelot

Go to the [wiki](https://github.com/shankj3/ocelot/wiki) for documentation and architecture. 

Ocelot is a distributed CI for running in container orchestration environments. It utilizes Vault, Consul, Postgres and NSQ and comes with a bangin' cli.


Future big wants:
- kuberentes werker nodes that interact with kube api 
- vagrant werker nodes
- github integration 

## Prometheus exports:
- ocelot_regex_failures
- ocelot_build_clean_failed
- ocelot_docker_api_errors_total
- ocelot_active_builds
- ocelot_build_duration_seconds
- ocelot_build_count_total
- ocelot_received_messages
- ocelot_werker_stream_errors_total
- ocelot_failed_cred
- ocelot_bitbucket_failed_calls
- ocelot_admin_request_proc_time
- ocelot_admin_active_requests
- admin_triggered_builds
- ocelot_recieved_hooks
- ocelot_db_active_requests
- ocelot_db_transaction_duration
