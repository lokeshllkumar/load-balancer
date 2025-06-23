#!/bin/bash
set -e

SERVICE_REGISTRY_DIR="spring-boot-service-registry"

if ! command -v mvn &> /dev/null
then
    echo "Error: Maven not found in PATH"
    exit 1
fi

cd "${SERVICE_REGISTRY_DIR}"

echo "Building the Service Registry..."
mvn clean install

echo "Built the Spring Boot Service Registry"