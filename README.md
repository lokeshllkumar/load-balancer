# http-load-balancer

An HTTP load balancer built using Go and the Gin framework, with support for multiple balancing strategies(round-robin, least connections, and sticky sessions).

## Features

- Balancing algorithms
    - Round-robin
    - Least connections
    - Sticky sessions
- Backend Health Checks - Monitors backend health at regular intervals
- TLS Encryption - Secure communication between the clients and the load balancer
- Metrics Monitoring - Tracks incoming requestsand backend performance

## Setup and Installation

- Prerequisites
    - Go 1.20+
    - OpenSSL(to generate certificates)
- Clone the Repository
```bash
git clone https://github.com/lokeshllkumar/load-balancer.git
cd load-balancer
```
- Generate TLS Certificates
    - Create a directory called ```certs``` to store the certificate
    - Generate an RSA key
    ```bash
    openssl gen rsa -out key.pem 2048
    ```
    - Generate a CSR file to create a Certificate
    ```
    openssl req -new -key key.pem -out server.csr
    ```
    - Generate the Self-Signed Certificate
    ```bash
    openssl x509 -req -days 365 -in server.csr -signkey key.pem -out cert.pem
    ```
    - Verify the Certificate
    ```bash
    openssl x509 -in cert.pem -text -noout
    ```
- Build the Project
```bash
go build -o load-balancer .
```
- Run the server
```bash
./load-balancer
```

## Usage

- Health Check Endpoint
    - Check if the server is running
    ```bash
    curl https://localhost:843/health --insecure
    ```
- Add a Backend Server
    - Add a backend server dynamically
    ```bash
    curl -X POST https://localhost:8443/api/backends -d '{"url": "https://backend3:8082"}' -H "Content-Type: application/json" --insecure
    ```
- List All Backends
    ```bash
    curl https://localhost:8443/api/backends --insecure
    ```

## Load Balancing Strategies

- Round Robin
    - Distributes requests evenly across all available backends.
- Least Connections
    - Directs traffic to the backend with the fewest active connections.
- Sticky Sessions
    - Ensures requests from the same client are directed to the same backend.

