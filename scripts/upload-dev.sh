#!/usr/bin/env bash

if [ -z "$1" ]; then
    echo "Usage: $0 <path> <adminkey>"
    exit 1
fi

DIR=$1
GAME=$(basename "$(dirname "$DIR")")
ZIPFILE="/tmp/${GAME}_$(date +%s).zip"

# Create temp directory for converted files
TMPDIR=$(mktemp -d --tmpdir=/tmp)
trap "rm -rf $TMPDIR" EXIT

# Convert TIF files to WebP
for tif in "$DIR"/*.tif "$DIR"/*.tiff; do
    [ -e "$tif" ] || continue
    filename=$(basename "${tif%.*}")
    cwebp -q 80 "$tif" -o "$TMPDIR/${filename}.webp"
done

# Copy JSON files
cp "$DIR"/*.json "$TMPDIR"/ 2>/dev/null
cp "$DIR"/*.webp "$TMPDIR"/ 2>/dev/null
cp "$DIR"/*.glb "$TMPDIR"/ 2>/dev/null

zip -rjq "$ZIPFILE" "$TMPDIR"
#stat "$ZIPFILE"
curl -H "Authorization: Bearer $2" \
        -X PUT http://localhost:8080/api/admin/import \
        -F "file=@\"${ZIPFILE}\""
rm -f "$ZIPFILE"