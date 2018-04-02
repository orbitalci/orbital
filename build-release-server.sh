#!/usr/bin/env sh
set -ex
# check for all the executables we need
command -v docker
if [ $? != 0 ]; then
  echo "need docker"
  exit 1
fi

command -v aws
if [ $? != 0 ]; then
  echo "need awscli"
  exit 1
fi

command -v zip
if [ $? != 0 ]; then
  echo "need zip"
  exit 1
fi


# on build server, we know the deps are already up to dat
#
#echo "building ocelot client"
## mac
#env GOOS=darwin GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
#zip -r mac-ocelot.zip ocelot
#rm ocelot
#
## window
#env GOOS=windows GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
#zip -r windows-ocelot.zip ocelot
#rm ocelot
#
## linux
#env GOOS=linux GOARCH=amd64 go build -o ocelot cmd/ocelot/main.go
#zip -r linux-ocelot.zip ocelot
#rm ocelot
#
#echo "uploading client binary"
#
## upload zipped client binary to s3
#aws s3 cp --acl public-read-write --content-disposition attachment mac-ocelot.zip s3://ocelotty/mac-ocelot.zip
#aws s3 cp --acl public-read-write --content-disposition attachment windows-ocelot.zip s3://ocelotty/windows-ocelot.zip
#aws s3 cp --acl public-read-write --content-disposition attachment linux-ocelot.zip s3://ocelotty/linux-ocelot.zip
#
#echo "uploading werker's template files"
#cd werker/builder/template
#tar -cvf werker_files.tar *
#
## upload zipped werker files to s3
#aws s3 cp --acl public-read-write --content-disposition attachment werker_files.tar s3://ocelotty/werker_files.tar
#
## cleanup the files we created for s3
#rm werker_files.tar
#cd -
#rm mac-ocelot.zip
#rm windows-ocelot.zip
#rm linux-ocelot.zip

docker tag jessishank/ocelot-build:kill.6 ocelot-build

echo "building admin"
docker build -f admin/Dockerfile -t jessishank/ocelot-admin:latest .
echo "building werker"
docker build -f werker/Dockerfile -t jessishank/ocelot-werker:latest .
echo "building poller"
docker build -f cmd/poller/Dockerfile -t jessishank/ocelot-poller:latest .
echo "building hookhandler"
docker build -f hookhandler/Dockerfile -t jessishank/ocelot-hookhandler:latest .

docker push jessishank/ocelot-admin:latest
docker push jessishank/ocelot-werker:latest
docker push jessishank/ocelot-hookhandler:latest
docker push jessishank/ocelot-poller:latest

echo "finished building for build server."
