#!/usr/bin/env sh


# make sure that all of our dependencies are up to date
dep ensure -v

echo "building go project"
# TODO: this only builds mac binary right now - swap to building other ones when we need it
env GOOS=darwin GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go

echo "uploading client binary"
# zip up the client binary
zip -r ocelot.zip ocelot

# upload zipped client binary into s3
aws s3 cp --acl public-read-write --content-disposition attachment ocelot.zip s3://ocelotty/ocelot.zip

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
