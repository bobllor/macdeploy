#!/usr/bin/env bash

binary_name="deploy"

env GOOS=darwin GOARCH=arm64 go build -o "$binary_name".bin