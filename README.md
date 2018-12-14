# OrbitalCI

OrbitalCI is a self-hostable continuous integration system.

All Continuous Integration systems are essentially fancy script executors. OrbitalCI is no different. What makes OrbitalCI different from other continuous integration systems?

* OrbitalCI is a container-first builder as a strategy for reproducing builds w/o cache implicitly affecting the failure or success.
* Users interact with OrbitalCI via command line interface. They can build, watch live logs and view history of their repo builds.
* Build environments and build instructions are laid out in a yaml configuration file that lives in the root of your code repo.
* Other self-hosted or private infrastructure (such as artifact repositories or Slack organizations) are supported in your builds.

The [wiki](https://github.com/level11consulting/orbitalci/wiki) has more in-depth documentation.

Table of contents:
- [OrbitalCI](#orbitalci)
  - [Roadmap](#roadmap)
  - [Developers](#developers)
    - [Getting started with Vagrant](#getting-started-with-vagrant)
      - [Requirements on host](#requirements-on-host)
        - [Known issues with Vagrant environment](#known-issues-with-vagrant-environment)
          - [vagrant up --provision fails on subsequent runs](#vagrant-up---provision-fails-on-subsequent-runs)
      - [Validate connectivity on host](#validate-connectivity-on-host)
    - [Getting started with Docker](#getting-started-with-docker)
      - [Manual setup steps w/ Docker](#manual-setup-steps-w-docker)
  - [Contributing](#contributing)

## Roadmap

OrbitalCI's roadmap is located [here](roadmap.md)

## Developers
### Getting started with Vagrant
#### Requirements on host
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

Updates to this process are welcome and should also be introduced via Github issue using `RFC` label. 