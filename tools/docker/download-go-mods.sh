#!/bin/bash
set -e

deps=($(go mod graph | awk '{if ($1 !~ "@") print $2}'))
for i in "${deps[@]}"
do
    go get "$i"
done
