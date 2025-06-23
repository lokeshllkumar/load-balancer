#!/bin/bash
set -e

LOAD_BALANCER_DIR="load-balancer"
LOAD_BALANCER_PROTO_DIR="proto"

if ! command -v go &> /dev/null
then
    echo "Error: Go not found in PATH"
    exit 1
fi

if ! command -v protoc &> /dev/null
then
    echo "Error: 'protoc' command not found"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null || ! command -v protoc-gen-go-grpc &> /dev/null
then
    echo "Error: Go gRPC plugins were not found in PATH"
    exit 1
fi

cd "${LOAD_BALANCER_DIR}"

cd "${LOAD_BALANCER_PROTO_DIR}"

protoc --go_out="." --go_opt=paths=source_relative \
       --go-grpc_out="." --go_opt=paths=source_relative \
       "${FLUX_PROTO_FILE}"

go mod tidy

echo "Building the Load Balancer executable..."
go build -o load-balancer

echo "The Load Balancer executable has been built"