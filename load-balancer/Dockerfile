# Go base image
FROM golang:1.22-alpine as builder

# set the working directory
WORKDIR /app

# install dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /load-balancer ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /load-balancer .

# configured to run on :8080 by default
EXPOSE 8080

CMD ["./load-balancer"]
