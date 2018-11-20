#!/usr/bin/env bash


PG_HOST=10.1.70.173
PG_PORT=5432
PG_USER=postgres
PG_PASSWORD=mysecretpassword

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
docker run --rm -v ${DIR}/sql:/flyway/sql boxfuse/flyway migrate -url=jdbc:postgresql://${PG_HOST}:${PG_PORT}/${PG_USER} -user=${PG_USER} -password=${PG_PASSWORD} -baselineOnMigrate=true