#!/bin/sh


# ===============
cd protos 
./build-protos.sh
cd ..
# ===============
cd werker
./build-protos.sh
cd ..
# ===============
# ===============
cd admin
./build-protos.sh
cd ..
# ===============

