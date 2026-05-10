#!/usr/bin/env sh
set -eu

mkdir -p "$PGDATA"
chown -R postgres:postgres /var/lib/postgresql/data

exec docker-entrypoint.sh postgres -c listen_addresses='*'
