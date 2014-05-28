#!/bin/sh

for package in $(go list ./...)
do
    DIR=$(echo $package | sed 's%github.com/materials-commons/base%.%')
    mkdir -p docs/$DIR
    godoc $package > docs/$DIR/package.txt
done
