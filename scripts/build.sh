#!/usr/bin/env bash
clear
echo "Building citrixadc-backup"
echo "-------------------------"
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

cd "$DIR"

echo "Cleaning up previous builds and packages"
rm -rf bin/*
rm -rf pkg/*


echo "Build executables per platform"
OUTPUT="bin/linux/amd64/citrixadc-backup"
echo " - linux-amd64 --> $OUTPUT"
GOOS=linux GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="bin/windows/amd64/citrixadc-backup.exe"
echo " - windows-amd64 --> $OUTPUT"
GOOS=windows GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="bin/darwin/amd64/citrixadc-backup"
echo " - darwin-amd64 --> $OUTPUT"
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT main.go

OUTPUT="bin/darwin/arm64/citrixadc-backup"
echo " - darwin-arm64 --> $OUTPUT"
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT main.go