#!/bin/sh
set -eu
db=${DATABASE_PATH:-services/freedom-bounties-api/data/freedom-bounties.db}
case "$db" in *.db) ;; *) echo "Refusing to remove a path that is not a .db file" >&2; exit 1;; esac
rm -f "$db" "$db-shm" "$db-wal"
echo "Demo database reset. It will be seeded on next API startup."
