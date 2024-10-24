#build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o ./load-balancer

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /build/load-balancer ./load-balancer

EXPOSE 8443

CMD ["/app/load-balancer"]
