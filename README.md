# load-balancer

A simple *Load Balancing and Service Discovery* system, capable of discovering and routing traffic to backend services registered with a *Service Registry* via a *Load Balancer*, with *Prometheus* observability integrated. Built simultaneously with [Flux](https://www.github.com/lokeshllkumar/flux), a Go module to build compatible backend services that can be registered with the Service Registry and deployed with the load balancer.

## Architecture

The project employs a microservices architectural pattern as described below:
- A <strong>Go</strong> *Load Balancer* acting as the entry point, distributing incoming requests amongst the registered instances of the backend service(s)
- A <strong>Spring Boot</strong> *Service Registry* serving as central hub for backend service(s) to register themselves and for the load balancer to discover them
- <strong>Prometheus</strong> for collecting metrics from the components to provide basic observability 
- Optional <strong>Grafana</strong> integration to visualize the metrics scraped by Prometheus with a custom dashboard

## Features

- Service Discovery: Backend services automatically and deregister with the service register upon spin-up and spin-down, respectively
- Health Checks: The load balancer peridiocally checks the health of registered backends
- Load Balancing Strategies: Provides support for a few load balancing strategies, namely
    - Round Robin
    - Least Connections
    - Sticky Sessions (keeps client bound to the same backend across several connection requests)
- Protocol Agnostic Registry Client - The load balancer and backend services can use either HTTP/REST or gRPC to communicate with the service registry
- Prometheus Metrics - A few key operational metrics are exposed by the Go services and the service registry, ready for scraping by Prometheus
- Visualzing Metrics - With the help of a Grafana dashboard, the metrics scraped by Prometheus can be visualized with the a plethora of graphs and charts

## Getting Started

- Prerequisites
    - Go
    - JDK 17+
    - Maven
    - protoc
    - Prometheus
    - Grafana (optional)
- Clone the repository
```bash
git clone https://github.com/lokeshllkumar/load-balancer
cd load-balancer
```
- Generate protobuf stubs (for gRPC). Ensure that ```protoc``` and Go gRPC plugins (```protoc-gen-go``` and ```protoc-gen-go-grpc```) are installed and in your PATH
```bash
# from within load-balancer/internal/proto/
cd load-balancer
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=../ --go-grpc_opt=paths=source_relative \
       service_registry.proto
```
> stubs are generated automatically during the build process for the Service Registry using the ```protbuf-maven-plugin```
- Build the Service Registry
```bash
cd service-registry
mvn clean package -Dskiptests
```
- Build the Load Balancer
```bash
cd load-balancer
go mod tidy
go build -o load-balancer .
```
- Build your Go Backend Services
```bash
# create a backend service either within this project's directory or elsewhere using the Flux module
cd backend-service
go build -o backend-service .
```

## Usage

- Start the Service Registry. It starts on port ```8081``` for HTTP and ```9091``` for gRPC by default, and can be configured to run on a different port. Prometheus metrics are accessible at ```/actuator/prometheus``` on port ```8081```
```bash
cd service-registry
java -jar target/service-registry-0.0.1-SNAPSHOT.jar
```
- Start the Backend Service Instances with the following environment variables set for each
```bash
cd backend-service
export SERVICE_NAME="backend-instance-1" # 2, 3, 4, ....
export PORT="8082" # the port must be unique to each instance
export METRICS_PORT="9092" # the port must be unique to each instance as well
export REGISTRY_URL="service-registry-address" # the reachable address of the service registry
export HOSTNAME="valid-host-name" # the service must be accessible via this host name
./backend-service

# can also run a service built with a different technology stack (can use the .proto file to generate protobuf stubs in the language of your choice to enable service regsitry communication via gRPC or define logic to leverage the REST API)
```
- Configure and start Prometheus
    - Define your Prometheus configuation [here](prometheus/prometheus.yml)
    - Start the Prometheus, which will run on port ```9090```
```bash
cd ~/prometheus
./prometheus --config.file=<path-to-your-prometheus-config-file>
```
- Configure and start the Load Balancer
    - Set the configuration of the Load Balancer [here](/load-balancer/config.yaml)
    - Start the Load Balancer, which starts on port ```8080``` by default, but can be configured on a different port of your choice.
```bash
cd load-balancer
./load-balancer
```
- Start the Grafana Server and create and add Prometheus as a data source. Create a new Dashboard Panel and in the Query tab, select the Prometheus data source added previously and pick a metric to visualize such as ```flux_registry_calls_total```, for example. You can play around with the wide variety of graphs and charts to visualize the various metrics.

