FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ip-json .
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/ip-json /usr/local/bin/ip-json
EXPOSE 8080
ENTRYPOINT ["ip-json"]
