#!/bin/sh
rm -f ./build/.env
cp .env ./build/.env
go mod download
go mod tidy
go build -o ./build/main ./cmd/api/main.go
