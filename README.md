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

# Developers
## Getting started with Vagrant
(Needs testing, more detail)
From `deploy` directory:
`vagrant up`

This will start 2 VMs. One for infra, one for ocelot components.
(Tip: Automatically sync code from host environment into VMs with `vagrant rsync-auto` in another terminal window after running `vagrant up`)

The ocelot VM will be at IP: `192.168.12.34`
The infra VM will be at IP: `192.168.56.78`

Infrastructure components run as Docker containers. (docker-compose files in `deploy/infra/`)
Consul UI: http://192.168.56.78:8500
Vault UI: http://192.168.56.78:8200 - Default token value: `ocelotdev`
NSQAdmin UI: http://192.168.56.78:4147
Postgres: 192.168.56.78:5432 - User name/Database name: `postgres`, Password: `mysecretpassword`

When the Ocelot VM starts up, it will instantiate the infrastructure by running `scripts/setup-cv.sh`, and attempt to install the current codebase.

You can ssh into these VMs
`vagrant ssh` & `vagrant ssh ocelot` will result in SSHing into the ocelot VM
`vagrant ssh infra` will SSH into the infra VM

The codebase will be located in `/home/vagrant/go/src/github.com/shankj3/ocelot`

### Validate connectivity on host
To configure the `ocelot` cli to point at this development instance, you need to set these environment variables:

`export ADMIN_HOST=192.168.12.34`
`export ADMIN_PORT=10000`

## Getting started with Docker
(Needs testing, more detail)
From the root of the repo:
`make docker-base`
`make docker-build`
(Start the infrastructure containers - minimum: nsq, consul, vault, postgres)
`docker-compose up`

## Contributing 

### Interfaces / mocks 

If you update any of the interfaces, you will also have to update the mocks generated by gomock and used for testing. Install [gomock](https://github.com/golang/mock) and then from root run `go generate ./...` 