#!/usr/bin/env sh


# make sure that all of our dependencies are up to date
dep ensure -v

# admin host should be set here
export ADMIN_HOST=ec2-34-212-13-136.us-west-2.compute.amazonaws.com

echo "building go project"
# build binary (RIGHT NOW WILL ONLY BUILD FOR MAC)
env GOOS=linux GOARCH=amd64 go install ./...

echo "uploading client binary"
# zip up the client binary
cd $HOME/go/bin
zip -r ocelot.zip ocelot

# upload zipped client binary into s3
aws s3 cp --acl public-read-write --content-disposition attachment ocelot.zip s3://ocelotty/ocelot.zip

# go back to original directory since we're going to build and stuff
cd -

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
