#!/usr/bin/env bash

vault write auth/approle/role/ocelot policies="ocelot,werker"