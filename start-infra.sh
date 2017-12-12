#!/bin/bash

docker-compose -f infra-docker-compose.yml -d up
# TODO: TJ has some stuff that will unseal vault for us! Perhaps it goes here?