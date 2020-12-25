Hermes can be built from source on Linux, Windows, or macOS.

## Prerequisites
- Git
- Golang 1.15 or newer

## Building
```shell
git clone https://github.com/c16a/hermes.git
cd hermes
go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app
```

### Cross compiling
To cross compile the Hermes binary to a different architecture or operating system, 
the `GOOS` and `GOARCH` environment variables can be used.
```shell
# List all available os/arch combinations for cross compiling
go tool dist list

# To compile the binary for Linux ARM 64-bit, use the below
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o binary_amd64 github.com/c16a/hermes/app
```
