#!/usr/bin/env sh


# make sure that all of our dependencies are up to date
dep ensure -v

echo "building ocelot client"
# mac
env GOOS=darwin GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
zip -r mac-ocelot.zip ocelot
rm ocelot

# window
env GOOS=windows GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
zip -r windows-ocelot.zip ocelot
rm ocelot

# linux
env GOOS=linux GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
zip -r linux-ocelot.zip ocelot
rm ocelot

# linux
pushd cmd/werker/
env GOOS=linux GOARCH=amd64 go build -o werker main.go
zip -r ../../linux-werker.zip werker
rm werker
popd


echo "uploading client binary"

# upload zipped client binary to s3
aws s3 cp --acl public-read-write --content-disposition attachment mac-ocelot.zip s3://ocelotty/mac-ocelot.zip
aws s3 cp --acl public-read-write --content-disposition attachment windows-ocelot.zip s3://ocelotty/windows-ocelot.zip
aws s3 cp --acl public-read-write --content-disposition attachment linux-ocelot.zip s3://ocelotty/linux-ocelot.zip
aws s3 cp --acl public-read-write --content-disposition attachment linux-werker.zip s3://ocelotty/linux-werker.zip

echo "uploading werker's template files"
cd werker/builder/template
tar -cvf werker_files.tar *

# upload zipped werker files to s3
aws s3 cp --acl public-read-write --content-disposition attachment werker_files.tar s3://ocelotty/werker_files.tar

# cleanup the files we created for s3
rm werker_files.tar
cd -
rm mac-ocelot.zip
rm windows-ocelot.zip
rm linux-ocelot.zip
rm linux-werker.zip


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
