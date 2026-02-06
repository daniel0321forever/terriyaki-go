# Build the application from source
FROM golang:latest

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
