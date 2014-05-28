#!/bin/sh

rm -rf test_data/.materials
rm -rf test_data/corrupted
rm -rf test_data/conversion
rm -f /tmp/sqltest*.db
mkdir -p test_data/.materials/projectdb
mkdir -p test_data/conversion/.materials
cp test_data/*.project test_data/.materials/projectdb
cp test_data/projects test_data/conversion/.materials/projects
cp test_data/.user test_data/.materials
mkdir -p /tmp/tproj/a
touch /tmp/tproj/a/a.txt

export MATERIALS_WEBDIR=""
export MATERIALS_ADDRESS=""
export MATERIALS_PORT=""
export MATERIALS_UPDATE_CHECK_INTERVAL=""
export MCDOWNLOADURL=""
export MCAPIURL=""
export MCURL=""

godep go test -v ./...
rm -rf /tmp/tproj

