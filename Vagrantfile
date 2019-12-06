# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu1804"

  # Set static IP for this host  
  config.vm.network "private_network", ip: "192.168.12.34"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus = "4"
  end

  # Use the vagrant provisioner to install Docker
  config.vm.provision "docker" do |d|
  end

  config.vm.synced_folder ".", "/home/vagrant/orbitalci",
    type: "virtualbox",
    disabled: false

  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    # Hack. Add in google's nameserver in resolv.conf everytime we log in with vagrant. Duplicates possible.
    echo 'echo "nameserver 8.8.8.8" | sudo tee -a /etc/resolv.conf' | tee -a ~/.bashrc
    sudo apt-get update
    sudo apt-get install -y curl git pkg-config libssl-dev build-essential

    # Install docker-compose via apt
    sudo apt install -y docker-compose

    # Install vault via wget
    wget https://releases.hashicorp.com/vault/1.3.0/vault_1.3.0_linux_amd64.zip
    unzip vault*.zip
    sudo mv vault /usr/local/bin
    rm vault*.zip
    # Set some env vars for future interactive sessions
    echo "export VAULT_ADDR=http://127.0.0.1:8200" | tee -a ~/.bashrc
    echo "export VAULT_TOKEN=orbital" | tee -a ~/.bashrc

    # Install rust via rustup script
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    ## Configure current session to add cargo to path
    source $HOME/.cargo/env
    ## Compile orb
    pushd orbitalci
    make

    # Install diesel cli w/ postgres extensions
    cargo install diesel_cli --no-default-features --features postgres

    # On startup, set diesel cli's MIGRATION_DIRECTORY env var to the postgres migration directory
    echo "export MIGRATION_DIRECTORY=~/orbitalci/models/orbital_database/postgres/migrations" | tee -a ~/.bashrc

    # On startup, set diesel cli's DATABASE_URL env var to the postgres connection string
    echo "export DATABASE_URL=postgres://orbital:orbital@localhost:5432/orbital" | tee -a ~/.bashrc

    # Start infrastructure services via docker-compose
    docker-compose -f docker-compose.infra.yml up -d
  SHELL
end