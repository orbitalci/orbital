#!/usr/bin/env sh


# make sure that all of our dependencies are up to date
dep ensure -v

# admin host should be set here
export ADMIN_HOST=ec2-34-212-13-136.us-west-2.compute.amazonaws.com

# build some binaries
go install ./...

# copy the client binary into s3
aws s3 cp --content-type=application/octet-stream $HOME/go/bin/ocelot s3://ocelotty/ocelot

# This build assumes you have your ssh key added to L11 bitbucket
# We need ssh keys to clone from the private bitbucket.

if [ -f ${SSH_PRIVATE_KEY:=${HOME}/.ssh/id_rsa} ]; then
   echo "Using private key: ${SSH_PRIVATE_KEY}"
else
   echo "Private key ${SSH_PRIVATE_KEY} not found. Set SSH_PRIVATE_KEY to your private key path"
   exit 1
fi

# Build this image first to cache dependencies
# add --squash when this is released into stable v of docker
docker build \
   --build-arg SSH_PRIVATE_KEY="$(cat ${SSH_PRIVATE_KEY})" \
   -f Dockerfile.build \
   -t ocelot-build \
   .

docker-compose build
