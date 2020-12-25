---
layout: page
title: Building from source
nav_order: 1
has_children: false
parent: Running locally
permalink: /running-locally/source
---

# Building from source
Hermes can be build on Linux, macOS, or Windows host systems.

## Prerequisites
- Git
- Golang 1.15 or above

## Building
You can build a binary for the current host system like below
```shell
git clone https://github.com/c16a/hermes
cd hermes
go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app
```

### Cross compiling
If you wish to cross compile for a different operating system or architecture,
you can use the `GOOS` and `GOARCH` environment variables.
```shell
# List all possible cross compilation combinations
go tool dist list

# For example, to build for Linux ARM 64-bit targets, use the below
GOOS=linux GOOS=arm64 go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app
```
