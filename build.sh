#!/bin/sh

# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o bin/caldavreport-linux

# macOS (Intel & ARM)
GOOS=darwin GOARCH=amd64 go build -o bin/caldavreport-macos-intel
GOOS=darwin GOARCH=arm64 go build -o bin/caldavreport-macos-arm

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o bin/caldavreport.exe

echo "Build completed for Linux, macOS, and Windows."