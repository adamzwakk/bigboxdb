#!/usr/bin/env bash

if [ -z "$1" ]; then
    echo "Usage: $0 <path> <adminkey>"
    exit 1
fi

for DIR in $1; do
    [ -d "$DIR" ] || continue
    GAME=$(basename "$(dirname "$DIR")")
    ZIPFILE="/tmp/${GAME}_$(date +%s).zip"
    
    # Create temp directory for converted files
    TMPDIR=$(mktemp -d --tmpdir=/tmp)
    
    # Convert TIF files to WebP
    for tif in "$DIR"/*.tif "$DIR"/*.tiff; do
        [ -e "$tif" ] || continue
        filename=$(basename "${tif%.*}")
        [ ! -f "$DIR/${filename}.webp" ] || continue
        cwebp -q 80 "$tif" -o "$TMPDIR/${filename}.webp"
        cp "$TMPDIR/${filename}.webp" "$DIR"
    done
    
    # Copy JSON and existing WebP files
    cp "$DIR"/*.json "$TMPDIR"/ 2>/dev/null
    cp "$DIR"/*.webp "$TMPDIR"/ 2>/dev/null
    cp "$DIR"/*.glb "$TMPDIR"/ 2>/dev/null
    
    zip -rjq "$ZIPFILE" "$TMPDIR"
    curl -H "Authorization: Bearer $2" \
        -X PUT https://www.bigboxdb.com/api/admin/import \
        -F "file=@\"${ZIPFILE}\""
    
    rm -f "$ZIPFILE"
    rm -rf "$TMPDIR"
done