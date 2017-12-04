#!/usr/bin/env sh

# Build this image first to cache dependencies
docker build -f Dockerfile.build -t ocelot-build .
docker-compose build
