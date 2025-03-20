FROM golang:1.23-alpine AS builder

# Build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o /app/noteo -ldflags="-w -s" .

FROM alpine:3.19

COPY --from=builder /app/noteo /noteo

ENTRYPOINT ["/noteo"]
