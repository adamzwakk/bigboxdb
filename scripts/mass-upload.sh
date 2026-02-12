#!/usr/bin/env bash

if [ -z "$1" ]; then
    echo "Usage: $0 <path> <adminkey>"
    exit 1
fi

for DIR in $1; do
    GAME=$(basename "$(dirname "$DIR")")
    ZIPFILE="/tmp/${GAME}_$(date +%s).zip"
    zip -rjq "$ZIPFILE" "$DIR"
    curl -H "Authorization: Bearer $2" \
         -X PUT http://localhost:8081/api/admin/import \
         -F "file=@\"${ZIPFILE}\""
    rm -f "$ZIPFILE"
done