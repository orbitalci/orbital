WIP

Start with vagrant: vagrant up

Start with docker-compose
Start the infra containers first

Manually fix vault to use v1
    export VAULT_ADDR=http://0.0.0.0:8200
    export VAULT_TOKEN=orbital
    vault secrets disable secret
    vault secrets enable -path=secret -version=1 kv

Run setup-cv
    export CONSUL_HTTP_ADDR=http://0.0.0.0:8500
    export VAULT_ADDR=http://0.0.0.0:8200
    export VAULT_TOKEN=orbital
    export DBHOST=0.0.0.0
    ./scripts/setup-cv.sh

# project ocelot

Go to the [wiki](https://github.com/level11consulting/ocelot/wiki) for documentation and architecture.

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
- ocelot_db_sqllib_error

# Developers
## Getting started with Vagrant

### Requirements on host
The following tools need to be installed on your host.

* Vagrant
* Virtualbox
  * If you use a different Vagrant provider, you may need to set your `VAGRANT_DEFAULT_PROVIDER` environment variable.
  * See [the Vagrant docs](https://www.vagrantup.com/docs/providers/default.html) for more detail
* Ansible 2.8+

From `deploy` directory:
`vagrant up`

This will start 2 VMs. One for infra, one for ocelot components.
(Tip: Automatically sync code from host environment into VMs with `vagrant rsync-auto` in another terminal window after running `vagrant up`)

The ocelot VM will be at IP: `192.168.12.34`

The infra VM will be at IP: `192.168.56.78`

Infrastructure components run as Docker containers. (docker-compose files in `deploy/infra/`)

* Consul UI: http://192.168.56.78:8500
* Vault UI: http://192.168.56.78:8200 - Default token value: `orbital`
* NSQAdmin UI: http://192.168.56.78:4171
* Postgres: 192.168.56.78:5432 - User name/Database name: `postgres`, Password: `mysecretpassword`

When the Vagrant VMs starts up, it will call use Ansible on your host to instantiate the Infra and Ocelot VMs, and attempt to install the current codebase.

You can ssh into these VMs
`vagrant ssh` & `vagrant ssh ocelot` will result in SSHing into the ocelot VM
`vagrant ssh infra` will SSH into the infra VM

The codebase will be located in `/home/vagrant/orbitalci`

OrbitalCI components are configured via Ansible, and start as systemd services so you can use `systemctl` or `journalctl` to interact.
* `orbital-admin`
* `orbital-hookhandler`
* `orbital-poller`
* `orbital-worker`

##### Known issues with Vagrant environment
###### vagrant up --provision fails on subsequent runs
```
TASK [vagrant_common : Install python modules that Ansible requires] ***********
changed: [infra] => (item=python-consul)
changed: [infra] => (item=docker)
changed: [infra] => (item=docker-compose)

TASK [vagrant_common : Download docker-compose] ********************************
```

For unknown reasons, on a subsequent pip install of `docker-compose`, the installation segfaults and hangs the provision.

Workaround:

`vagrant destroy` to delete the vms, and then recreate the VMs with `vagrant up`.

#### Validate connectivity on host
To configure the `ocelot` cli to point at this development instance, you need to set these environment variables:

* `export ADMIN_HOST=192.168.12.34`
* `export ADMIN_PORT=10000`

### Getting started with Docker
From the root of the repo:

1. `make docker-build` # To build local docker images
2. `make start-docker-infra` # Start the infrastructure containers
3. Follow the manual steps below
4. `make start-docker-orbital` # Start the orbital services

#### Manual setup steps w/ Docker
These steps require `consul` and `vault` to be installed on your host.

Manually fix vault to use v1

    export VAULT_ADDR=http://localhost:8200
    export VAULT_TOKEN=orbital
    vault secrets disable secret
    vault secrets enable -path=secret -version=1 kv

Run setup-cv

    export CONSUL_HTTP_ADDR=http://localhost:8500
    export VAULT_ADDR=http://localhost:8200
    export VAULT_TOKEN=orbital
    export DBHOST=db
    ./scripts/setup-cv.sh

Load database schema

    export PG_HOST=db
    export PG_PORT=5432
    export PG_USER=postgres
    export PG_PASSWORD=mysecretpassword
    docker run --network orbital --rm -v $(pwd)/deploy/sql:/flyway/sql boxfuse/flyway migrate -url=jdbc:postgresql://${PG_HOST}:${PG_PORT}/${PG_USER} -user=${PG_USER} -password=${PG_PASSWORD} -baselineOnMigrate=true 

## Contributing 

Fork the repo and issue a pull request.

Inspired by the [Rust governance process](https://www.rust-lang.org/governance), large or potentially backwards incompatible changes should be socialized by opening an issue with the `RFC` label.

Conversation aims to mainly occur in the Github issues. This process is being developed to provide transparency of intent when decisions are made. The desire is to allow stakeholders time to engage with objections, but have process to continue moving forward if there is a lack of response within a reasonable time period.

When an RFC's details are worked out and ready for final comment, similar but less formal to Rust's governance, `RFC` issues, the issue opener, or the maintainers should call out in the comments stating final changes are in.

At this point, more changes may be required due to new comments. But after 10 calendar days (at least 5 business days), code or additional design based on this RFC is acceptable.

If you update any of the interfaces, you will also have to update the mocks generated by gomock and used for testing. Install [gomock](https://github.com/golang/mock) and then from root run `go generate ./...` 