# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu1804"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus = "2"
  end

  config.vm.synced_folder ".", "/home/vagrant/orbitalci",
    type: "virtualbox",
    disabled: false

  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    sudo apt-get update
    sudo apt-get install -y curl git pkg-config libssl-dev build-essential

    # Install docker via get docker script
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker vagrant

    # Install rust via rustup script
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    ## Configure current session to add cargo to path
    source $HOME/.cargo/env
    ## Compile orb
    pushd orbitalci
    make
  SHELL
end