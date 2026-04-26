#!/bin/sh
set -xe

rm -f ./build/.env

if [ -f .env ]; then
	cp .env ./build/.env
fi

go mod download
go mod tidy
go build -o ./build/main ./cmd/api/main.go
