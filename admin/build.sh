#!/bin/bash

protoc -I models/ models/guideocelot.proto --go_out=plugins=grpc:models
