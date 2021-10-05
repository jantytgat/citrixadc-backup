#!/usr/bin/env bash
clear
echo "Cleaning up citrixadc-backup"
echo "----------------------------"
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

cd "$DIR"

echo "Cleaning up previous builds and packages"
rm -rf bin
rm -rf pkg