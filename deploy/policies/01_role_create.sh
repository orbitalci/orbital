#!/usr/bin/env bash
echo "writing ocelot policy"
vault policy write ocelot ocelot.hcl
echo "writing werker policy"
vault policy write werker werker_deployer.hcl
