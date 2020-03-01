# OrbitalCI

OrbitalCI is a self-hostable continuous integration system written in Rust.

All Continuous Integration systems are essentially fancy script executors. OrbitalCI is no different. What makes OrbitalCI different from other continuous integration systems?

* OrbitalCI is a container-first builder as a strategy for reproducing builds w/o cache implicitly affecting the failure or success.
* Users interact with OrbitalCI via command line interface. They can build, watch live logs and view history of their repo builds.
* Build environments and build instructions are laid out in a yaml configuration file that lives in the root of your code repo.
* Other self-hosted or private infrastructure (such as artifact repositories or Slack organizations) are supported in your builds.

Table of contents:
- [OrbitalCI](#OrbitalCI)
  - [Roadmap](#Roadmap)
  - [Developers](#Developers)
    - [Requirements on host](#Requirements-on-host)
    - [Getting started with Vagrant](#Getting-started-with-Vagrant)
      - [Requirements on host](#Requirements-on-host-1)
    - [Getting started with Docker](#Getting-started-with-Docker)
  - [Contributing](#Contributing)

## Roadmap

OrbitalCI's roadmap for the year is located [here](roadmap.md)

## Developers

To get started, just run `make`.

### Requirements on host
* Rust 1.39+
* Docker 
* make
* git

Note: Docker container exec only works on Linux hosts due to https://github.com/softprops/shiplift/issues/155

### Getting started with Vagrant
#### Requirements on host
The following tools need to be installed on your host.

* Vagrant
* Virtualbox
  * If you use a different Vagrant provider, you may need to set your `VAGRANT_DEFAULT_PROVIDER` environment variable.
  * See [the Vagrant docs](https://www.vagrantup.com/docs/providers/default.html) for more detail

From root of the repo:
`vagrant up`

The codebase will share/sync the current directory in the VM under `/home/vagrant/orbitalci`.

### Getting started with Docker

This image is not yet ready for active usage, but it can be used for manual testing of the cli through `orb dev [...]` commands.

To manually build the container from root of the repo:
`docker build -t orb .`

## Contributing 

Fork the repo and issue a pull request.